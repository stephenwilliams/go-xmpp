package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
)

const xmppStreamOpen = "<?xml version='1.0'?><stream:stream to='%s' xmlns='%s' xmlns:stream='%s' version='1.0'>"

type Session struct {
	// Session info
	BindJid      string // Jabber ID as provided by XMPP server
	StreamId     string
	Features     streamFeatures
	TlsEnabled   bool
	lastPacketId int

	// Session interface
	In  chan interface{}
	Out chan interface{}

	// read / write
	socketProxy io.ReadWriter // TODO Rename
	decoder     *xml.Decoder

	// Config
	config *Config

	// error management
	err error
}

func NewSession(conn net.Conn, o Config) (net.Conn, *Session, error) {
	s := &Session{
		config: &o,
	}
	s.init(conn, o)

	// starttls
	var tlsConn net.Conn
	tlsConn = s.startTlsIfSupported(conn, o.parsedJid.Domain)
	if s.TlsEnabled {
		s.reset(conn, tlsConn, o)
	}

	if !s.TlsEnabled && !o.Insecure {
		return nil, nil, errors.New("failed to negotiate TLS")
	}

	// auth
	s.auth(o)
	s.reset(tlsConn, tlsConn, o)

	// bind Resource and 'start' XMPP session
	s.bind(o)
	s.rfc3921Session(o)

	return tlsConn, s, s.err
}

func (s *Session) PacketId() string {
	s.lastPacketId++
	return fmt.Sprintf("%x", s.lastPacketId)
}

func (s *Session) init(conn net.Conn, o Config) {
	s.socketProxy = conn
	s.Features = s.open(o.parsedJid.Domain)
}

func (s *Session) reset(conn net.Conn, newConn net.Conn, o Config) {
	if s.err != nil {
		return
	}

	s.Features = s.open(o.parsedJid.Domain)
}

func (s *Session) open(domain string) (f streamFeatures) {
	// Send stream open tag
	if _, s.err = printf(s.config, s.socketProxy, xmppStreamOpen, domain, NSClient, NSStream); s.err != nil {
		return
	}

	// Set xml decoder and extract streamID from reply
	s.StreamId, s.err = initDecoder(s.decoder) // TODO refactor / rename
	if s.err != nil {
		return
	}

	// extract stream features
	if s.err = s.decoder.Decode(&f); s.err != nil {
		s.err = errors.New("stream open decode features: " + s.err.Error())
	}
	return
}

func (s *Session) startTlsIfSupported(conn net.Conn, domain string) net.Conn {
	if s.err != nil {
		return conn
	}

	if s.Features.StartTLS.XMLName.Space+" "+s.Features.StartTLS.XMLName.Local == nsTLS+" starttls" {
		printf(s.config, s.socketProxy, "<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>")

		var k tlsProceed
		if s.err = s.decoder.DecodeElement(&k, nil); s.err != nil {
			s.err = errors.New("expecting starttls proceed: " + s.err.Error())
			return conn
		}
		s.TlsEnabled = true

		// TODO: add option to accept all TLS certificates: insecureSkipTlsVerify (DefaultTlsConfig.InsecureSkipVerify)
		DefaultTlsConfig.ServerName = domain
		var tlsConn *tls.Conn = tls.Client(conn, &DefaultTlsConfig)
		// We convert existing connection to TLS
		if s.err = tlsConn.Handshake(); s.err != nil {
			return tlsConn
		}

		// We check that cert matches hostname
		s.err = tlsConn.VerifyHostname(domain)
		return tlsConn
	}

	// starttls is not supported => we do not upgrade the connection:
	return conn
}

func (s *Session) auth(o Config) {
	if s.err != nil {
		return
	}

	s.err = authSASL(s.socketProxy, s.decoder, s.Features, o.parsedJid.Username, o.Password)
}

func (s *Session) bind(o Config) {
	if s.err != nil {
		return
	}

	// Send IQ message asking to bind to the local user name.
	var resource = o.parsedJid.Resource
	if resource != "" {
		printf(s.config, s.socketProxy, "<iq type='set' id='%s'><bind xmlns='%s'><Resource>%s</Resource></bind></iq>",
			s.PacketId(), nsBind, resource)
	} else {
		printf(s.config, s.socketProxy, "<iq type='set' id='%s'><bind xmlns='%s'/></iq>", s.PacketId(), nsBind)
	}

	var iq IQ
	if s.err = s.decoder.Decode(&iq); s.err != nil {
		s.err = errors.New("error decoding iq bind result: " + s.err.Error())
		return
	}

	// TODO Check all elements
	switch payload := iq.Payload[0].(type) {
	case *BindBind:
		s.BindJid = payload.Jid // our local id (with possibly randomly generated Resource
	default:
		s.err = errors.New("iq bind result missing")
	}

	return
}

// TODO: remove when ejabberd is fixed: https://github.com/processone/ejabberd/issues/869
// After the bind, if the session is required (as per old RFC 3921), we send the session open iq
func (s *Session) rfc3921Session(o Config) {
	if s.err != nil {
		return
	}

	var iq IQ
	if s.Features.Session.optional.Local != "" {
		printf(s.config, s.socketProxy, "<iq type='set' id='%s'><session xmlns='%s'/></iq>", s.PacketId(), nsSession)
		if s.err = s.decoder.Decode(&iq); s.err != nil {
			s.err = errors.New("expecting iq result after session open: " + s.err.Error())
			return
		}
	}
}

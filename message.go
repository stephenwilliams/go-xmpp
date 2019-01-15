package xmpp

import (
	"encoding/xml"
	"fmt"
)

type ChatState string

const (
	ChatStateActive    ChatState = "active"
	ChatStateComposing ChatState = "composing"
	ChatStatePaused    ChatState = "paused"
	ChatStateInactive  ChatState = "inactive"
	ChatStateGone      ChatState = "gone"
)

// ============================================================================
// Message Packet

type Message struct {
	XMLName xml.Name `xml:"message"`
	PacketAttrs
	Subject   string    `xml:"subject,omitempty"`
	Body      string    `xml:"body,omitempty"`
	Thread    string    `xml:"thread,omitempty"`
	Error     Err       `xml:"error,omitempty"`
	ChatState ChatState `xml:",omitempty"`
	AMP       *AMP      `xml:",omitempty"`
}

func (Message) Name() string {
	return "message"
}

func NewMessage(msgtype, from, to, id, lang string) Message {
	return Message{
		XMLName: xml.Name{Local: "message"},
		PacketAttrs: PacketAttrs{
			Id:   id,
			From: from,
			To:   to,
			Type: msgtype,
			Lang: lang,
		},
	}
}

type messageDecoder struct{}

var message messageDecoder

func (messageDecoder) decode(p *xml.Decoder, se xml.StartElement) (Message, error) {
	var packet Message
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

func (m *Message) XMPPFormat() string {
	return fmt.Sprintf("<message to='%s' type='chat' xml:lang='en'>"+
		"<body>%s</body></message>",
		m.To,
		xmlEscape(m.Body))
}

func (c ChatState) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if c == "" {
		return nil
	}

	start.Name.Local = string(c)
	start.Name.Space = "http://jabber.org/protocol/chatstates"

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	return e.EncodeToken(start.End())
}

func (m *Message) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	m.XMLName = start.Name
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			m.Id = attr.Value
		case "from":
			m.From = attr.Value
		case "to":
			m.To = attr.Value
		case "type":
			m.Type = attr.Value
		case "lang":
			m.Lang = attr.Value
		}
	}

	var curr string
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Space == "http://jabber.org/protocol/chatstates" {
				m.ChatState = ChatState(t.Name.Local)
				continue
			}

			switch t.Name.Local {
			case "body":
				curr = "body"
			case "thread":
				curr = "thread"
			case "subject":
				curr = "subject"
			case "error":
				e := &Err{}
				err := d.DecodeElement(e, &t)
				if err != nil {
					return err
				}
				m.Error = *e
			}
		case xml.CharData:
			s := string([]byte(t))
			switch curr {
			case "body":
				m.Body = s
			case "thread":
				m.Thread = s
			case "subject":
				m.Subject = s
			}
			curr = ""
		case xml.EndElement:
			if t == start.End() {
				return nil
			}
		}
	}
}

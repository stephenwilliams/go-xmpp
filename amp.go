package xmpp

import "encoding/xml"

type AMP struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/amp amp"`
	Status  string   `xml:"status,omitempty"`
	Rules   []*AMPRule
}

type AMPRule struct {
	XMLName   xml.Name `xml:"rule"`
	Condition string   `xml:"condition,omitempty"`
	Value     string   `xml:"value,omitempty"`
	Action    string   `xml:"action,omitempty"`
}

func (a AMP) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if a.Status != "" {
		start.Attr = append(start.Attr, xml.Attr{
			Name: xml.Name{
				Local: "status",
			},
			Value: a.Status,
		})
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for _, rule := range a.Rules {
		err = e.Encode(rule)
		if err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (a *AMP) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	a.XMLName = start.Name

	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "status":
			a.Status = attr.Value
		}
	}

	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch t := t.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "rule":
				r := &AMPRule{}
				err := d.DecodeElement(r, &t)
				if err != nil {
					return err
				}
				a.Rules = append(a.Rules, r)
			}
		case xml.EndElement:
			if t == start.End() {
				return nil
			}
		}
	}
}

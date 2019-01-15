package xmpp

import (
	"encoding/xml"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestAMP_XML(t *testing.T) {
	amp := AMP{
		Status: "test",
		Rules: []*AMPRule{
			{
				Condition: "receipt",
				Value:     "received",
				Action:    "notify",
			},
		},
	}

	data, err := xml.Marshal(amp)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	parsedAMP := AMP{}
	if err = xml.Unmarshal(data, &parsedAMP); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedAMP, amp) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedAMP, amp))
	}
}

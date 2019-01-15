package xmpp

import (
	"testing"
)

func TestValidJids(t *testing.T) {
	var jid *Jid
	var err error

	goodJids := []string{"test@Domain.com", "test@Domain.com/Resource"}

	for i, sjid := range goodJids {
		if jid, err = NewJid(sjid); err != nil {
			t.Error("could not parse correct jid")
		}

		if jid.Username != "test" {
			t.Error("incorrect jid Username")
		}

		if jid.Domain != "Domain.com" {
			t.Error("incorrect jid Domain")
		}

		if i == 0 && jid.Resource != "" {
			t.Error("bare jid Resource should be empty")
		}

		if i == 1 && jid.Resource != "Resource" {
			t.Error("incorrect full jid Resource")
		}
	}
}

// TODO: Check if Resource cannot contain a /
func TestIncorrectJids(t *testing.T) {
	badJids := []string{"test@Domain.com@otherdomain.com",
		"test@Domain.com/test/test"}

	for _, sjid := range badJids {
		if _, err := NewJid(sjid); err == nil {
			t.Error("parsing incorrect jid should return error: " + sjid)
		}
	}
}

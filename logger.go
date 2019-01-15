package xmpp

import (
	"fmt"
	"io"
)

func print(c *Config, w io.Writer, a ...interface{}) (int, error) {
	s := fmt.Sprint(a...)

	if c.PacketLogger != nil {
		c.PacketLogger.LogSend(s)
	}

	return fmt.Fprint(w, s)
}

func printf(c *Config, w io.Writer, f string, a ...interface{}) (int, error) {
	s := fmt.Sprintf(f, a...)

	if c.PacketLogger != nil {
		c.PacketLogger.LogSend(s)
	}

	return fmt.Fprint(w, s)
}

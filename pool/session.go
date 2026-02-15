//go:build linux

package pool

import (
	"github.com/amosdavis/pool-go/poolioc"
)

// SessionState returns the human-readable session state.
func (c *Conn) SessionState() (string, error) {
	info, err := c.SessionInfo()
	if err != nil {
		return "", err
	}
	return stateString(info.State), nil
}

func stateString(s uint8) string {
	switch s {
	case poolioc.StateIdle:
		return "IDLE"
	case poolioc.StateInitSent:
		return "INIT_SENT"
	case poolioc.StateChallenged:
		return "CHALLENGED"
	case poolioc.StateEstablished:
		return "ESTABLISHED"
	case poolioc.StateRekeying:
		return "REKEYING"
	case poolioc.StateClosing:
		return "CLOSING"
	default:
		return "UNKNOWN"
	}
}

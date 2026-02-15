//go:build linux

package pool

import (
	"net"
	"sync"
	"time"

	"github.com/amosdavis/pool-go/poolioc"
)

// Conn implements [net.Conn] over a POOL session.
//
// All methods are safe for concurrent use. Read and Write operate on
// the default channel (0). Use [Conn.OpenChannel] for multi-channel I/O.
type Conn struct {
	dev        *poolioc.Device
	sessionIdx uint32
	localAddr  *Addr
	remoteAddr *Addr
	channel    uint8

	mu           sync.Mutex
	closed       bool
	readDeadline  time.Time
	writeDeadline time.Time
}

// newConn creates a Conn from an established session.
func newConn(dev *poolioc.Device, idx uint32, local, remote *Addr, ch uint8) *Conn {
	return &Conn{
		dev:        dev,
		sessionIdx: idx,
		localAddr:  local,
		remoteAddr: remote,
		channel:    ch,
	}
}

// Read reads data from the POOL session.
// It implements [io.Reader].
func (c *Conn) Read(b []byte) (int, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return 0, ErrClosed
	}
	deadline := c.readDeadline
	c.mu.Unlock()

	if len(b) == 0 {
		return 0, nil
	}

	type result struct {
		n   int
		err error
	}

	ch := make(chan result, 1)
	go func() {
		n, err := c.dev.RecvBytes(c.sessionIdx, c.channel, b)
		ch <- result{n, err}
	}()

	if !deadline.IsZero() {
		timer := time.NewTimer(time.Until(deadline))
		defer timer.Stop()

		select {
		case r := <-ch:
			return r.n, mapErrno(r.err)
		case <-timer.C:
			return 0, &timeoutError{}
		}
	}

	r := <-ch
	return r.n, mapErrno(r.err)
}

// Write writes data to the POOL session.
// It implements [io.Writer].
func (c *Conn) Write(b []byte) (int, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return 0, ErrClosed
	}
	deadline := c.writeDeadline
	c.mu.Unlock()

	if len(b) == 0 {
		return 0, nil
	}

	if len(b) > poolioc.MaxPayload {
		return 0, ErrMessageTooLarge
	}

	type result struct {
		err error
	}

	ch := make(chan result, 1)
	go func() {
		err := c.dev.SendBytes(c.sessionIdx, c.channel, b)
		ch <- result{err}
	}()

	if !deadline.IsZero() {
		timer := time.NewTimer(time.Until(deadline))
		defer timer.Stop()

		select {
		case r := <-ch:
			if r.err != nil {
				return 0, mapErrno(r.err)
			}
			return len(b), nil
		case <-timer.C:
			return 0, &timeoutError{}
		}
	}

	r := <-ch
	if r.err != nil {
		return 0, mapErrno(r.err)
	}
	return len(b), nil
}

// Close closes the POOL session.
// It implements [io.Closer].
func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrClosed
	}
	c.closed = true

	return mapErrno(c.dev.CloseSession(c.sessionIdx))
}

// LocalAddr returns the local address.
func (c *Conn) LocalAddr() net.Addr {
	if c.localAddr == nil {
		return &Addr{}
	}
	return c.localAddr
}

// RemoteAddr returns the remote peer address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// SetDeadline sets both read and write deadlines.
func (c *Conn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDeadline = t
	c.writeDeadline = t
	return nil
}

// SetReadDeadline sets the deadline for Read calls.
func (c *Conn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDeadline = t
	return nil
}

// SetWriteDeadline sets the deadline for Write calls.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDeadline = t
	return nil
}

// SessionIndex returns the kernel session index.
func (c *Conn) SessionIndex() uint32 {
	return c.sessionIdx
}

// Telemetry returns the latest telemetry for this session.
// The telemetry is refreshed by the kernel every HeartbeatSec seconds.
func (c *Conn) Telemetry() (*poolioc.Telemetry, error) {
	sessions, err := c.dev.Sessions()
	if err != nil {
		return nil, mapErrno(err)
	}
	for i := range sessions {
		if sessions[i].Index == c.sessionIdx {
			t := sessions[i].Telem
			return &t, nil
		}
	}
	return nil, ErrNotEstablished
}

// SessionInfo returns detailed session information.
func (c *Conn) SessionInfo() (*poolioc.SessionInfo, error) {
	sessions, err := c.dev.Sessions()
	if err != nil {
		return nil, mapErrno(err)
	}
	for i := range sessions {
		if sessions[i].Index == c.sessionIdx {
			info := sessions[i]
			return &info, nil
		}
	}
	return nil, ErrNotEstablished
}

// Verify interface compliance at compile time.
var _ net.Conn = (*Conn)(nil)

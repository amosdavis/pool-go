//go:build linux

package pool

import (
	"net"
	"sync"
	"time"

	"github.com/amosdavis/pool-go/poolioc"
)

// ChannelConn wraps a [Conn] to operate on a specific POOL channel.
// It implements [net.Conn]. The channel is subscribed on creation and
// unsubscribed on Close.
type ChannelConn struct {
	conn    *Conn
	channel uint8

	mu     sync.Mutex
	closed bool
}

// OpenChannel subscribes to a channel on an existing Conn and returns
// a [ChannelConn] for reading and writing on that channel.
func (c *Conn) OpenChannel(channel uint8) (*ChannelConn, error) {
	if err := c.dev.ChannelSubscribe(c.sessionIdx, channel); err != nil {
		return nil, mapErrno(err)
	}
	return &ChannelConn{
		conn:    c,
		channel: channel,
	}, nil
}

// Read reads data from this channel.
func (cc *ChannelConn) Read(b []byte) (int, error) {
	cc.mu.Lock()
	if cc.closed {
		cc.mu.Unlock()
		return 0, ErrClosed
	}
	cc.mu.Unlock()

	n, err := cc.conn.dev.RecvBytes(cc.conn.sessionIdx, cc.channel, b)
	return n, mapErrno(err)
}

// Write writes data to this channel.
func (cc *ChannelConn) Write(b []byte) (int, error) {
	cc.mu.Lock()
	if cc.closed {
		cc.mu.Unlock()
		return 0, ErrClosed
	}
	cc.mu.Unlock()

	if len(b) > poolioc.MaxPayload {
		return 0, ErrMessageTooLarge
	}

	if err := cc.conn.dev.SendBytes(cc.conn.sessionIdx, cc.channel, b); err != nil {
		return 0, mapErrno(err)
	}
	return len(b), nil
}

// Close unsubscribes from the channel. The underlying session is NOT closed.
func (cc *ChannelConn) Close() error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.closed {
		return ErrClosed
	}
	cc.closed = true

	return mapErrno(cc.conn.dev.ChannelUnsubscribe(cc.conn.sessionIdx, cc.channel))
}

// LocalAddr returns the local address.
func (cc *ChannelConn) LocalAddr() net.Addr  { return cc.conn.LocalAddr() }

// RemoteAddr returns the remote peer address.
func (cc *ChannelConn) RemoteAddr() net.Addr { return cc.conn.RemoteAddr() }

// SetDeadline is not supported on ChannelConn; use the parent Conn.
func (cc *ChannelConn) SetDeadline(t time.Time) error      { return nil }

// SetReadDeadline is not supported on ChannelConn; use the parent Conn.
func (cc *ChannelConn) SetReadDeadline(t time.Time) error   { return nil }

// SetWriteDeadline is not supported on ChannelConn; use the parent Conn.
func (cc *ChannelConn) SetWriteDeadline(t time.Time) error  { return nil }

// Channel returns the channel number.
func (cc *ChannelConn) Channel() uint8 { return cc.channel }

// Verify interface compliance at compile time.
var _ net.Conn = (*ChannelConn)(nil)

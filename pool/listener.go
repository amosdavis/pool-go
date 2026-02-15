//go:build linux

package pool

import (
	"net"
	"sync"
	"time"

	"github.com/amosdavis/pool-go/poolioc"
)

// Listener implements [net.Listener] for the POOL protocol.
//
// Call [Listen] to create a Listener. Then call Accept to wait for
// incoming POOL sessions.
type Listener struct {
	dev     *poolioc.Device
	addr    *Addr
	mu      sync.Mutex
	closed  bool
	known   map[uint32]struct{}
	pollInt time.Duration
}

// Listen starts listening for POOL connections on the given address.
// The network must be "pool", "pool4", or "pool6". The address is
// "host:port" or ":port".
func Listen(network, address string) (*Listener, error) {
	addr, err := ResolveAddr(network, address)
	if err != nil {
		// Allow ":port" shorthand with an unspecified host
		if address[0] == ':' {
			addr = &Addr{IP: net.IPv6unspecified, Port: 0}
			var portErr error
			addr.Port, portErr = net.LookupPort("tcp", address[1:])
			if portErr != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	dev, err := poolioc.Open()
	if err != nil {
		return nil, mapErrno(err)
	}

	if err := dev.Listen(uint16(addr.Port)); err != nil {
		dev.Close()
		return nil, mapErrno(err)
	}

	return &Listener{
		dev:     dev,
		addr:    addr,
		known:   make(map[uint32]struct{}),
		pollInt: 100 * time.Millisecond,
	}, nil
}

// Accept waits for and returns the next POOL connection.
// It polls the kernel session list for newly established sessions.
func (l *Listener) Accept() (net.Conn, error) {
	for {
		l.mu.Lock()
		if l.closed {
			l.mu.Unlock()
			return nil, ErrClosed
		}
		l.mu.Unlock()

		sessions, err := l.dev.Sessions()
		if err != nil {
			return nil, mapErrno(err)
		}

		for i := range sessions {
			s := &sessions[i]
			if s.State != poolioc.StateEstablished {
				continue
			}
			l.mu.Lock()
			_, already := l.known[s.Index]
			if !already {
				l.known[s.Index] = struct{}{}
			}
			l.mu.Unlock()
			if already {
				continue
			}

			remote := &Addr{
				IP:   net.IP(s.PeerAddr[:]).To16(),
				Port: int(s.PeerPort),
			}
			return newConn(l.dev, s.Index, l.addr, remote, 0), nil
		}

		time.Sleep(l.pollInt)
	}
}

// Close stops the POOL listener and releases the device.
func (l *Listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return ErrClosed
	}
	l.closed = true

	err := l.dev.Stop()
	_ = l.dev.Close()
	return mapErrno(err)
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.addr
}

// Verify interface compliance at compile time.
var _ net.Listener = (*Listener)(nil)

//go:build linux

package pool

import (
	"fmt"
	"net"
	"syscall"
)

// Addr represents a POOL endpoint address.
// It implements [net.Addr].
type Addr struct {
	IP   net.IP
	Port int
}

// Network returns "pool".
func (a *Addr) Network() string { return "pool" }

// String returns the address formatted as "host:port".
// IPv6 addresses use bracket notation.
func (a *Addr) String() string {
	if a == nil {
		return "<nil>"
	}
	return net.JoinHostPort(a.IP.String(), fmt.Sprintf("%d", a.Port))
}

// AddrFamily returns syscall.AF_INET or syscall.AF_INET6.
func (a *Addr) AddrFamily() uint8 {
	if a.IP.To4() != nil {
		return syscall.AF_INET
	}
	return syscall.AF_INET6
}

// PeerAddrBytes returns the 16-byte address suitable for ConnectReq.PeerAddr.
// IPv4 addresses are returned as IPv4-mapped IPv6.
func (a *Addr) PeerAddrBytes() [16]byte {
	var out [16]byte
	ip16 := a.IP.To16()
	if ip16 != nil {
		copy(out[:], ip16)
	}
	return out
}

// ResolveAddr parses an address string into an Addr.
// The address can be "host:port" where host is an IPv4 address, an IPv6
// address (with or without brackets), or a hostname.
func ResolveAddr(network, address string) (*Addr, error) {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("pool: invalid address %q: %w", address, err)
	}

	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		return nil, fmt.Errorf("pool: invalid port %q: %w", portStr, err)
	}

	ip := net.ParseIP(host)
	if ip == nil {
		// Hostname — resolve it
		addrs, err := net.LookupIP(host)
		if err != nil {
			return nil, fmt.Errorf("pool: cannot resolve %q: %w", host, err)
		}
		if len(addrs) == 0 {
			return nil, fmt.Errorf("pool: no addresses for %q", host)
		}
		ip = addrs[0]
	}

	return &Addr{IP: ip, Port: port}, nil
}

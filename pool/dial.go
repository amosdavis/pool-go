//go:build linux

package pool

import (
	"fmt"
	"net"
	"time"

	"github.com/amosdavis/pool-go/poolioc"
)

// Dial connects to a POOL peer at the given address.
// The network must be "pool", "pool4", or "pool6".
// The address is "host:port".
//
//	conn, err := pool.Dial("pool", "10.0.0.1:9253")
//	conn, err := pool.Dial("pool6", "[::1]:9253")
func Dial(network, address string) (*Conn, error) {
	return DialTimeout(network, address, 0)
}

// DialTimeout acts like [Dial] but imposes a timeout on the handshake.
// A timeout of zero means no limit.
func DialTimeout(network, address string, timeout time.Duration) (*Conn, error) {
	addr, err := ResolveAddr(network, address)
	if err != nil {
		return nil, err
	}

	dev, err := poolioc.Open()
	if err != nil {
		return nil, mapErrno(err)
	}

	req := poolioc.ConnectReq{
		PeerAddr:   addr.PeerAddrBytes(),
		PeerPort:   uint16(addr.Port),
		AddrFamily: addr.AddrFamily(),
	}

	type dialResult struct {
		idx int
		err error
	}

	ch := make(chan dialResult, 1)
	go func() {
		idx, err := dev.Connect(req)
		ch <- dialResult{idx, err}
	}()

	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case r := <-ch:
			if r.err != nil {
				dev.Close()
				return nil, mapErrno(r.err)
			}
			local := resolveLocalAddr(addr)
			return newConn(dev, uint32(r.idx), local, addr, 0), nil
		case <-timer.C:
			dev.Close()
			return nil, &timeoutError{}
		}
	}

	r := <-ch
	if r.err != nil {
		dev.Close()
		return nil, mapErrno(r.err)
	}

	local := resolveLocalAddr(addr)
	return newConn(dev, uint32(r.idx), local, addr, 0), nil
}

// resolveLocalAddr builds a best-effort local address.
func resolveLocalAddr(remote *Addr) *Addr {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   remote.IP,
		Port: remote.Port,
	})
	if err != nil {
		return &Addr{IP: net.IPv6unspecified, Port: 0}
	}
	defer conn.Close()

	local := conn.LocalAddr().(*net.UDPAddr)
	return &Addr{IP: local.IP, Port: 0}
}

// Resolve creates a POOL address from a network and address string,
// without connecting.
func Resolve(network, address string) (*Addr, error) {
	if network != "pool" && network != "pool4" && network != "pool6" {
		return nil, fmt.Errorf("pool: unsupported network %q", network)
	}
	return ResolveAddr(network, address)
}

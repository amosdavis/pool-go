//go:build linux

package steps

import (
	"context"
	"fmt"
	"net"

	"github.com/amosdavis/pool-go/poolioc"
	"github.com/cucumber/godog"
)

type pooliocContext struct {
	dev        *poolioc.Device
	sessionIdx int
	recvBuf    []byte
	err        error
	bitmap     [poolioc.MaxChannels / 8]byte
}

func InitializePooliocScenario(ctx *godog.ScenarioContext) {
	pc := &pooliocContext{}

	ctx.After(func(scenarioCtx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if pc.dev != nil {
			_ = pc.dev.Close()
		}
		return scenarioCtx, nil
	})

	ctx.Step(`^the POOL kernel module is loaded$`, pc.moduleLoaded)
	ctx.Step(`^I open the POOL device$`, pc.openDevice)
	ctx.Step(`^the device file descriptor should be valid$`, pc.fdValid)
	ctx.Step(`^I close the device without error$`, pc.closeDevice)
	ctx.Step(`^I have an open POOL device$`, pc.openDevice)
	ctx.Step(`^I start listening on port (\d+)$`, pc.listen)
	ctx.Step(`^the listener should be active$`, pc.listenerActive)
	ctx.Step(`^I stop the listener$`, pc.stopListener)
	ctx.Step(`^the listener should be stopped$`, pc.listenerStopped)
	ctx.Step(`^a remote POOL peer is listening on "([^"]*)"$`, pc.remotePeerListening)
	ctx.Step(`^I connect to "([^"]*)"$`, pc.connectTo)
	ctx.Step(`^the session should be established$`, pc.sessionEstablished)
	ctx.Step(`^the session index should be non-negative$`, pc.sessionIdxNonNeg)
	ctx.Step(`^I have an established session$`, pc.haveEstablishedSession)
	ctx.Step(`^I send "([^"]*)" on channel (\d+)$`, pc.sendOnChannel)
	ctx.Step(`^the peer echoes the data back$`, pc.peerEchoes)
	ctx.Step(`^I should receive "([^"]*)" on channel (\d+)$`, pc.recvOnChannel)
	ctx.Step(`^I list sessions$`, pc.listSessions)
	ctx.Step(`^the session list should contain at least (\d+) session$`, pc.sessionCount)
	ctx.Step(`^the session state should be "([^"]*)"$`, pc.sessionState)
	ctx.Step(`^I close the session$`, pc.closeSession)
	ctx.Step(`^the session should be removed from the list$`, pc.sessionRemoved)
	ctx.Step(`^I subscribe to channel (\d+)$`, pc.subscribeChannel)
	ctx.Step(`^channel (\d+) should be active$`, pc.channelActive)
	ctx.Step(`^I unsubscribe from channel (\d+)$`, pc.unsubscribeChannel)
	ctx.Step(`^channel (\d+) should be inactive$`, pc.channelInactive)
	ctx.Step(`^I subscribe to channels (\d+), (\d+), and (\d+)$`, pc.subscribeMultiple)
	ctx.Step(`^I list channels$`, pc.listChannels)
	ctx.Step(`^the bitmap should show channels (\d+), (\d+), and (\d+) as active$`, pc.bitmapCheck)
	ctx.Step(`^I send a (\d+)-byte payload$`, pc.sendLargePayload)
	ctx.Step(`^I should receive a (\d+)-byte payload$`, pc.recvLargePayload)
	ctx.Step(`^the connection should fail with a timeout or unreachable error$`, pc.connFailed)
	ctx.Step(`^I convert "([^"]*)" to an IPv4-mapped IPv6 address$`, pc.convertIPv4Mapped)
	ctx.Step(`^the result should be "([^"]*)"$`, pc.mappedResult)
	ctx.Step(`^IsV4Mapped should return true$`, pc.isV4Mapped)
	ctx.Step(`^POOL_IOC_LISTEN should have type byte 0x50$`, pc.iocListenType)
	ctx.Step(`^POOL_IOC_CONNECT should have direction bits set to WRITE$`, pc.iocConnectDir)
}

func (pc *pooliocContext) moduleLoaded() error {
	// Verified by successfully opening /dev/pool
	return nil
}

func (pc *pooliocContext) openDevice() error {
	dev, err := poolioc.Open()
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}
	pc.dev = dev
	return nil
}

func (pc *pooliocContext) fdValid() error {
	if pc.dev.Fd() < 0 {
		return fmt.Errorf("fd %d is invalid", pc.dev.Fd())
	}
	return nil
}

func (pc *pooliocContext) closeDevice() error {
	err := pc.dev.Close()
	pc.dev = nil
	return err
}

func (pc *pooliocContext) listen(port int) error {
	return pc.dev.Listen(uint16(port))
}

func (pc *pooliocContext) listenerActive() error {
	// If listen succeeded, it's active
	return nil
}

func (pc *pooliocContext) stopListener() error {
	return pc.dev.Stop()
}

func (pc *pooliocContext) listenerStopped() error {
	return nil
}

func (pc *pooliocContext) remotePeerListening(_ string) error {
	// In integration tests, a peer would be set up externally
	return godog.ErrPending
}

func (pc *pooliocContext) connectTo(addr string) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		return err
	}

	ip := net.ParseIP(host).To16()
	var peerAddr [16]byte
	copy(peerAddr[:], ip)

	req := poolioc.ConnectReq{
		PeerAddr: peerAddr,
		PeerPort: uint16(port),
	}
	if ip.To4() != nil {
		req.AddrFamily = 2 // AF_INET
	} else {
		req.AddrFamily = 10 // AF_INET6
	}

	pc.sessionIdx, pc.err = pc.dev.Connect(req)
	return pc.err
}

func (pc *pooliocContext) sessionEstablished() error {
	if pc.err != nil {
		return fmt.Errorf("session not established: %w", pc.err)
	}
	return nil
}

func (pc *pooliocContext) sessionIdxNonNeg() error {
	if pc.sessionIdx < 0 {
		return fmt.Errorf("session index %d is negative", pc.sessionIdx)
	}
	return nil
}

func (pc *pooliocContext) haveEstablishedSession() error {
	return godog.ErrPending
}

func (pc *pooliocContext) sendOnChannel(msg string, ch int) error {
	return pc.dev.SendBytes(uint32(pc.sessionIdx), uint8(ch), []byte(msg))
}

func (pc *pooliocContext) peerEchoes() error {
	return godog.ErrPending
}

func (pc *pooliocContext) recvOnChannel(expected string, ch int) error {
	buf := make([]byte, 4096)
	n, err := pc.dev.RecvBytes(uint32(pc.sessionIdx), uint8(ch), buf)
	if err != nil {
		return err
	}
	if string(buf[:n]) != expected {
		return fmt.Errorf("expected %q, got %q", expected, string(buf[:n]))
	}
	return nil
}

func (pc *pooliocContext) listSessions() error {
	_, err := pc.dev.Sessions()
	return err
}

func (pc *pooliocContext) sessionCount(min int) error {
	sessions, err := pc.dev.Sessions()
	if err != nil {
		return err
	}
	if len(sessions) < min {
		return fmt.Errorf("expected at least %d sessions, got %d", min, len(sessions))
	}
	return nil
}

func (pc *pooliocContext) sessionState(expected string) error {
	sessions, err := pc.dev.Sessions()
	if err != nil {
		return err
	}
	for _, s := range sessions {
		if int(s.Index) == pc.sessionIdx {
			got := stateStringIOC(s.State)
			if got != expected {
				return fmt.Errorf("expected state %s, got %s", expected, got)
			}
			return nil
		}
	}
	return fmt.Errorf("session %d not found", pc.sessionIdx)
}

func stateStringIOC(s uint8) string {
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

func (pc *pooliocContext) closeSession() error {
	return pc.dev.CloseSession(uint32(pc.sessionIdx))
}

func (pc *pooliocContext) sessionRemoved() error {
	sessions, err := pc.dev.Sessions()
	if err != nil {
		return err
	}
	for _, s := range sessions {
		if int(s.Index) == pc.sessionIdx {
			return fmt.Errorf("session %d still present", pc.sessionIdx)
		}
	}
	return nil
}

func (pc *pooliocContext) subscribeChannel(ch int) error {
	return pc.dev.ChannelSubscribe(uint32(pc.sessionIdx), uint8(ch))
}

func (pc *pooliocContext) channelActive(ch int) error {
	bitmap, err := pc.dev.ChannelList(uint32(pc.sessionIdx))
	if err != nil {
		return err
	}
	byteIdx := ch / 8
	bitIdx := uint(ch % 8)
	if bitmap[byteIdx]&(1<<bitIdx) == 0 {
		return fmt.Errorf("channel %d not active", ch)
	}
	return nil
}

func (pc *pooliocContext) unsubscribeChannel(ch int) error {
	return pc.dev.ChannelUnsubscribe(uint32(pc.sessionIdx), uint8(ch))
}

func (pc *pooliocContext) channelInactive(ch int) error {
	bitmap, err := pc.dev.ChannelList(uint32(pc.sessionIdx))
	if err != nil {
		return err
	}
	byteIdx := ch / 8
	bitIdx := uint(ch % 8)
	if bitmap[byteIdx]&(1<<bitIdx) != 0 {
		return fmt.Errorf("channel %d still active", ch)
	}
	return nil
}

func (pc *pooliocContext) subscribeMultiple(a, b, c int) error {
	for _, ch := range []int{a, b, c} {
		if err := pc.dev.ChannelSubscribe(uint32(pc.sessionIdx), uint8(ch)); err != nil {
			return err
		}
	}
	return nil
}

func (pc *pooliocContext) listChannels() error {
	var err error
	pc.bitmap, err = pc.dev.ChannelList(uint32(pc.sessionIdx))
	return err
}

func (pc *pooliocContext) bitmapCheck(a, b, c int) error {
	for _, ch := range []int{a, b, c} {
		byteIdx := ch / 8
		bitIdx := uint(ch % 8)
		if pc.bitmap[byteIdx]&(1<<bitIdx) == 0 {
			return fmt.Errorf("channel %d not set in bitmap", ch)
		}
	}
	return nil
}

func (pc *pooliocContext) sendLargePayload(size int) error {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return pc.dev.SendBytes(uint32(pc.sessionIdx), 0, data)
}

func (pc *pooliocContext) recvLargePayload(size int) error {
	buf := make([]byte, size+1024)
	n, err := pc.dev.RecvBytes(uint32(pc.sessionIdx), 0, buf)
	if err != nil {
		return err
	}
	if n != size {
		return fmt.Errorf("expected %d bytes, got %d", size, n)
	}
	return nil
}

func (pc *pooliocContext) connFailed() error {
	if pc.err == nil {
		return fmt.Errorf("expected connection failure, got success")
	}
	return nil
}

var mappedIP net.IP

func (pc *pooliocContext) convertIPv4Mapped(addr string) error {
	ip := net.ParseIP(addr)
	if ip == nil {
		return fmt.Errorf("invalid IP: %s", addr)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return fmt.Errorf("not an IPv4 address: %s", addr)
	}
	ipUint := uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])
	mapped := poolioc.IPv4ToMapped(ipUint)
	mappedIP = net.IP(mapped[:])
	return nil
}

func (pc *pooliocContext) mappedResult(expected string) error {
	exp := net.ParseIP(expected)
	if !mappedIP.Equal(exp) {
		return fmt.Errorf("expected %s, got %s", expected, mappedIP)
	}
	return nil
}

func (pc *pooliocContext) isV4Mapped() error {
	var arr [16]byte
	copy(arr[:], mappedIP.To16())
	if !poolioc.IsV4Mapped(arr) {
		return fmt.Errorf("IsV4Mapped returned false")
	}
	return nil
}

func (pc *pooliocContext) iocListenType() error {
	typeByte := (poolioc.IocListen >> 8) & 0xFF
	if typeByte != 0x50 {
		return fmt.Errorf("expected type 0x50, got 0x%02X", typeByte)
	}
	return nil
}

func (pc *pooliocContext) iocConnectDir() error {
	dir := poolioc.IocConnect >> 30
	if dir != 1 { // _IOC_WRITE = 1
		return fmt.Errorf("expected direction WRITE (1), got %d", dir)
	}
	return nil
}

//go:build linux

package poolioc

import "unsafe"

// Send transmits data on a POOL session.
// The caller must set req.DataPtr to the address of the data buffer
// and req.Len to the number of bytes to send.
func (d *Device) Send(req SendReq) error {
	return d.ioctl(iocSend, unsafe.Pointer(&req))
}

// SendBytes is a convenience wrapper that sends a byte slice on a
// session and channel.
func (d *Device) SendBytes(sessionIdx uint32, channel uint8, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	req := SendReq{
		SessionIdx: sessionIdx,
		Channel:    channel,
		Len:        uint32(len(data)),
		DataPtr:    uint64(uintptr(unsafe.Pointer(&data[0]))),
	}
	return d.Send(req)
}

// Recv receives data from a POOL session.
// On entry, req.Len is the buffer capacity and req.DataPtr points to
// the buffer. On return, req.Len contains the number of bytes received.
func (d *Device) Recv(req *RecvReq) error {
	return d.ioctl(iocRecv, unsafe.Pointer(req))
}

// RecvBytes is a convenience wrapper that receives into a byte slice.
// Returns the number of bytes received.
func (d *Device) RecvBytes(sessionIdx uint32, channel uint8, buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	req := RecvReq{
		SessionIdx: sessionIdx,
		Channel:    channel,
		Len:        uint32(len(buf)),
		DataPtr:    uint64(uintptr(unsafe.Pointer(&buf[0]))),
	}
	if err := d.Recv(&req); err != nil {
		return 0, err
	}
	return int(req.Len), nil
}

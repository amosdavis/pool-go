//go:build linux

package poolioc

import "unsafe"

// Listen starts the POOL listener on the given port.
func (d *Device) Listen(port uint16) error {
	return d.ioctl(iocListen, unsafe.Pointer(&port))
}

// Connect initiates a POOL handshake to a remote peer.
// Returns the session index on success.
func (d *Device) Connect(req ConnectReq) (int, error) {
	return d.ioctlRet(iocConnect, unsafe.Pointer(&req))
}

// Stop shuts down the POOL listener.
func (d *Device) Stop() error {
	return d.ioctl(iocStop, unsafe.Pointer(nil))
}

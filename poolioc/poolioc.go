//go:build linux

package poolioc

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

const devicePath = "/dev/pool"

// Device represents an open handle to the POOL kernel module.
// All methods are safe for concurrent use.
type Device struct {
	mu sync.Mutex
	fd int
}

// Open opens /dev/pool and returns a Device handle.
func Open() (*Device, error) {
	fd, err := syscall.Open(devicePath, syscall.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("poolioc: open %s: %w", devicePath, err)
	}
	return &Device{fd: fd}, nil
}

// Close closes the device handle.
func (d *Device) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.fd < 0 {
		return os.ErrClosed
	}
	err := syscall.Close(d.fd)
	d.fd = -1
	return err
}

// Fd returns the underlying file descriptor. Returns -1 if closed.
func (d *Device) Fd() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.fd
}

// ioctl performs a raw ioctl syscall on the device fd.
func (d *Device) ioctl(req uintptr, arg unsafe.Pointer) error {
	d.mu.Lock()
	fd := d.fd
	d.mu.Unlock()

	if fd < 0 {
		return os.ErrClosed
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		req,
		uintptr(arg),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// ioctlRet performs an ioctl and returns the positive return value
// (used by CONNECT which returns the session index).
func (d *Device) ioctlRet(req uintptr, arg unsafe.Pointer) (int, error) {
	d.mu.Lock()
	fd := d.fd
	d.mu.Unlock()

	if fd < 0 {
		return -1, os.ErrClosed
	}

	r1, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		req,
		uintptr(arg),
	)
	if errno != 0 {
		return -1, errno
	}
	return int(r1), nil
}

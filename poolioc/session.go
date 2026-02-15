//go:build linux

package poolioc

import "unsafe"

// Sessions retrieves the list of active POOL sessions.
func (d *Device) Sessions() ([]SessionInfo, error) {
	infos := make([]SessionInfo, MaxSessions)
	list := SessionList{
		MaxSessions: MaxSessions,
		InfoPtr:     uint64(uintptr(unsafe.Pointer(&infos[0]))),
	}

	if err := d.ioctl(iocSessions, unsafe.Pointer(&list)); err != nil {
		return nil, err
	}

	return infos[:list.Count], nil
}

// CloseSession closes the session at the given index.
func (d *Device) CloseSession(idx uint32) error {
	return d.ioctl(iocCloseSess, unsafe.Pointer(&idx))
}

//go:build linux

package poolioc

import "unsafe"

// ChannelSubscribe subscribes to receive data on the given channel
// for the specified session.
func (d *Device) ChannelSubscribe(sessionIdx uint32, channel uint8) error {
	req := ChannelReq{
		SessionIdx: sessionIdx,
		Channel:    channel,
		Operation:  ChanSubscribe,
	}
	return d.ioctl(iocChannel, unsafe.Pointer(&req))
}

// ChannelUnsubscribe stops receiving data on the given channel.
func (d *Device) ChannelUnsubscribe(sessionIdx uint32, channel uint8) error {
	req := ChannelReq{
		SessionIdx: sessionIdx,
		Channel:    channel,
		Operation:  ChanUnsubscribe,
	}
	return d.ioctl(iocChannel, unsafe.Pointer(&req))
}

// ChannelList returns a 256-bit bitmap of active channels for the session.
// Bit i is set if channel i is subscribed.
func (d *Device) ChannelList(sessionIdx uint32) ([MaxChannels / 8]byte, error) {
	var bitmap [MaxChannels / 8]byte
	req := ChannelReq{
		SessionIdx: sessionIdx,
		Operation:  ChanList,
		DataPtr:    uint64(uintptr(unsafe.Pointer(&bitmap[0]))),
	}
	if err := d.ioctl(iocChannel, unsafe.Pointer(&req)); err != nil {
		return bitmap, err
	}
	return bitmap, nil
}

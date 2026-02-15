// Package poolioc provides low-level access to the POOL kernel module
// via ioctl calls on /dev/pool.
//
// This package mirrors the C structures and constants defined in pool.h
// and is intended for applications that need direct control over the
// POOL ioctl interface. For idiomatic Go networking, use the pool package
// which implements net.Conn and net.Listener on top of poolioc.
//
// This package requires Linux with the pool.ko kernel module loaded.
package poolioc

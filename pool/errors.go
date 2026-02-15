//go:build linux

package pool

import (
	"errors"
	"fmt"
	"syscall"
)

// Sentinel errors for POOL operations.
var (
	// ErrSessionFull indicates the kernel session table is full (ENOSPC).
	ErrSessionFull = errors.New("pool: session table full")

	// ErrAuthFailed indicates the POOL handshake authentication failed.
	ErrAuthFailed = errors.New("pool: authentication failed")

	// ErrClosed indicates the connection or listener has been closed.
	ErrClosed = errors.New("pool: connection closed")

	// ErrTimeout indicates a deadline was exceeded.
	ErrTimeout = errors.New("pool: operation timed out")

	// ErrMessageTooLarge indicates the message exceeds MaxPayload.
	ErrMessageTooLarge = errors.New("pool: message too large")

	// ErrBufferTooSmall indicates the receive buffer is too small (EMSGSIZE).
	ErrBufferTooSmall = errors.New("pool: buffer too small")

	// ErrNotEstablished indicates the session is not in ESTABLISHED state.
	ErrNotEstablished = errors.New("pool: session not established")

	// ErrNetUnreachable indicates the peer is unreachable.
	ErrNetUnreachable = errors.New("pool: network unreachable")
)

// mapErrno converts a syscall.Errno to a typed POOL error.
func mapErrno(err error) error {
	if err == nil {
		return nil
	}

	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return err
	}

	switch errno {
	case syscall.ENOSPC:
		return ErrSessionFull
	case syscall.ECONNREFUSED:
		return ErrAuthFailed
	case syscall.ETIMEDOUT:
		return ErrTimeout
	case syscall.EMSGSIZE:
		return ErrMessageTooLarge
	case syscall.ENETUNREACH:
		return ErrNetUnreachable
	case syscall.EBADF:
		return ErrClosed
	default:
		return fmt.Errorf("pool: %w", errno)
	}
}

// timeoutError implements net.Error for deadline exceeded.
type timeoutError struct{}

func (e *timeoutError) Error() string   { return ErrTimeout.Error() }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

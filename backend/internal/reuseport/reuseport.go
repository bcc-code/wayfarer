//go:build linux || darwin

// Package reuseport creates TCP listeners with SO_REUSEPORT, letting two
// server processes (blue/green deploys) bind the same port simultaneously:
// the kernel distributes new connections across all bound processes, so a
// deploy is start-green -> health-check -> stop-blue with zero refused
// connections and no proxy in front.
package reuseport

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// Listen returns a TCP listener on addr with SO_REUSEPORT set.
func Listen(ctx context.Context, addr string) (net.Listener, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var sockErr error
			err := c.Control(func(fd uintptr) {
				sockErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			})
			if err != nil {
				return err
			}
			return sockErr
		},
	}
	l, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("reuseport listen %s: %w", addr, err)
	}
	return l, nil
}

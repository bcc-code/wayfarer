//go:build linux || darwin

package reuseport

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListen_TwoProcessesCanShareAPort(t *testing.T) {
	ctx := context.Background()

	// First listener grabs an ephemeral port.
	l1, err := Listen(ctx, "127.0.0.1:0")
	require.NoError(t, err)
	defer l1.Close()

	addr := l1.Addr().String()

	// Second listener binds the SAME address — the blue/green overlap.
	l2, err := Listen(ctx, addr)
	require.NoError(t, err, "second reuseport bind on %s must succeed", addr)
	defer l2.Close()

	// A plain listener on that address must still fail.
	_, err = net.Listen("tcp", addr)
	assert.Error(t, err, "non-reuseport bind should conflict")
}

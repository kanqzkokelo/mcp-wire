package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/mcp-wire/mcp-wire/pkg/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestP2PMeshDialAndPing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	priv1, _ := identity.NewEphemeralIdentity()
	node1, err := NewHost(ctx, priv1, DefaultListenAddrs(0))
	require.NoError(t, err)
	defer node1.Close()

	priv2, _ := identity.NewEphemeralIdentity()
	node2, err := NewHost(ctx, priv2, DefaultListenAddrs(0))
	require.NoError(t, err)
	defer node2.Close()

	// Register stream handler on node1
	handlerCalled := make(chan bool, 1)
	node1.SetStreamHandler("/mcp-wire/v1/test-mesh", func(s network.Stream) {
		handlerCalled <- true
		s.Close()
	})

	// Dial node1 from node2
	s, err := DialStream(ctx, node2.Host, node1.AddrInfo(), "/mcp-wire/v1/test-mesh")
	require.NoError(t, err)
	_, err = s.Write([]byte("ping\n"))
	require.NoError(t, err)
	defer s.Close()

	select {
	case <-handlerCalled:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("stream handler was not called within timeout")
	}

	// Ping test
	rtt := PingPeer(ctx, node2.Host, node1.Host.ID())
	assert.True(t, rtt > 0, "expected positive RTT latency")
}

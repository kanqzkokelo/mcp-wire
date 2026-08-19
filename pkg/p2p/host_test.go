package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/mcp-wire/mcp-wire/pkg/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLibp2pHostInitialization(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	priv1, err := identity.NewEphemeralIdentity()
	require.NoError(t, err)

	node1, err := NewHost(ctx, priv1, DefaultListenAddrs(0))
	require.NoError(t, err)
	defer node1.Close()

	assert.NotEmpty(t, node1.Host.ID().String())

	addrs := GetFormattedAddrs(node1.Host)
	assert.NotEmpty(t, addrs)

	// Test connection between 2 hosts
	priv2, err := identity.NewEphemeralIdentity()
	require.NoError(t, err)

	node2, err := NewHost(ctx, priv2, DefaultListenAddrs(0))
	require.NoError(t, err)
	defer node2.Close()

	// Connect node2 to node1
	err = node2.Host.Connect(ctx, node1.AddrInfo())
	require.NoError(t, err)

	assert.Equal(t, node2.Host.Network().Connectedness(node1.Host.ID()).String(), "Connected")
}

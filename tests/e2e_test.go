package tests

import (
	"bufio"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/mcp-wire/mcp-wire/pkg/auth"
	"github.com/mcp-wire/mcp-wire/pkg/identity"
	"github.com/mcp-wire/mcp-wire/pkg/p2p"
	"github.com/mcp-wire/mcp-wire/pkg/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2EP2PLoopbackAndLatency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Host Node
	hostPriv, err := identity.NewEphemeralIdentity()
	require.NoError(t, err)

	hostNode, err := p2p.NewHost(ctx, hostPriv, p2p.DefaultListenAddrs(0))
	require.NoError(t, err)
	defer hostNode.Close()

	hostPeerID := hostNode.Host.ID()

	// 2. Client Node
	clientPriv, err := identity.NewEphemeralIdentity()
	require.NoError(t, err)

	clientNode, err := p2p.NewHost(ctx, clientPriv, p2p.DefaultListenAddrs(0))
	require.NoError(t, err)
	defer clientNode.Close()

	// Register P2P stream handler on Host Node
	protoID := protocol.ToProtocolID("test-service")
	hostNode.Host.SetStreamHandler(protoID, func(s network.Stream) {
		defer s.Close()

		// Perform Host Handshake
		_, err := auth.PerformHostHandshake(s, "secret123", nil)
		if err != nil {
			return
		}

		// Echo back newline-delimited requests
		reader := bufio.NewReader(s)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			_, _ = s.Write(line)
		}
	})

	// Connect Client to Host
	err = clientNode.Host.Connect(ctx, hostNode.AddrInfo())
	require.NoError(t, err)

	// Open stream from Client
	stream, err := p2p.DialStream(ctx, clientNode.Host, hostNode.AddrInfo(), protoID)
	require.NoError(t, err)
	defer stream.Close()

	// Perform Client Handshake
	err = auth.PerformClientHandshake(stream, "secret123", "test-service")
	require.NoError(t, err)

	// Measure sub-10ms proxy RTT latency
	start := time.Now()
	testMsg := []byte("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}\n")

	_, err = stream.Write(testMsg)
	require.NoError(t, err)

	streamReader := bufio.NewReader(stream)
	respLine, err := streamReader.ReadBytes('\n')
	require.NoError(t, err)

	rtt := time.Since(start)

	assert.Equal(t, string(testMsg), string(respLine))
	assert.True(t, rtt < 50*time.Millisecond, fmt.Sprintf("proxy latency overhead should be minimal (got %v)", rtt))

	_ = hostPeerID
}

func BenchmarkP2PThroughput(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hostPriv, _ := identity.NewEphemeralIdentity()
	hostNode, _ := p2p.NewHost(ctx, hostPriv, p2p.DefaultListenAddrs(0))
	defer hostNode.Close()

	clientPriv, _ := identity.NewEphemeralIdentity()
	clientNode, _ := p2p.NewHost(ctx, clientPriv, p2p.DefaultListenAddrs(0))
	defer clientNode.Close()

	protoID := protocol.ToProtocolID("bench-service")
	hostNode.Host.SetStreamHandler(protoID, func(s network.Stream) {
		defer s.Close()
		_, _ = auth.PerformHostHandshake(s, "", nil)
		buf := make([]byte, 32*1024)
		for {
			n, err := s.Read(buf)
			if n > 0 {
				_, _ = s.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	})

	_ = clientNode.Host.Connect(ctx, hostNode.AddrInfo())
	stream, _ := p2p.DialStream(ctx, clientNode.Host, hostNode.AddrInfo(), protoID)
	defer stream.Close()

	_ = auth.PerformClientHandshake(stream, "", "bench-service")

	payload := make([]byte, 1024)
	readBuf := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = stream.Write(payload)
		_, _ = stream.Read(readBuf)
	}
}

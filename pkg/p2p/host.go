package p2p

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
)

type Node struct {
	Host   host.Host
	Gater  *ConnectionGater
	Cancel context.CancelFunc
}

// DefaultListenAddrs returns TCP & QUIC listening multiaddrs for IPv4 & IPv6
func DefaultListenAddrs(port int) []string {
	if port <= 0 {
		port = 0
	}
	return []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", port),
		fmt.Sprintf("/ip6/::/tcp/%d", port),
		fmt.Sprintf("/ip6/::/udp/%d/quic-v1", port),
	}
}

// NewHost creates and initializes a full libp2p host node with Noise & Yamux
func NewHost(ctx context.Context, privKey crypto.PrivKey, listenAddrs []string) (*Node, error) {
	if len(listenAddrs) == 0 {
		listenAddrs = DefaultListenAddrs(0)
	}

	gater := NewConnectionGater()

	opts := []libp2p.Option{
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(listenAddrs...),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
		libp2p.ConnectionGater(gater),
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelayService(),
		libp2p.EnableAutoRelayWithStaticRelays(nil),
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize libp2p host: %w", err)
	}

	nodeCtx, cancel := context.WithCancel(ctx)
	ListenEvents(nodeCtx, h)

	slog.Info("libp2p host initialized successfully", "peer_id", h.ID().String())

	return &Node{
		Host:   h,
		Gater:  gater,
		Cancel: cancel,
	}, nil
}

// GetFormattedAddrs returns listening multiaddrs complete with /p2p/<PeerID>
func GetFormattedAddrs(h host.Host) []string {
	pid := h.ID().String()
	addrs := h.Addrs()
	res := make([]string, 0, len(addrs))
	for _, a := range addrs {
		res = append(res, fmt.Sprintf("%s/p2p/%s", a.String(), pid))
	}
	return res
}

// CloseHost gracefully closes all streams and shuts down host
func (n *Node) Close() error {
	if n.Cancel != nil {
		n.Cancel()
	}
	if n.Host != nil {
		return n.Host.Close()
	}
	return nil
}

// AddrInfo returns peer.AddrInfo for this node
func (n *Node) AddrInfo() peer.AddrInfo {
	return peer.AddrInfo{
		ID:    n.Host.ID(),
		Addrs: n.Host.Addrs(),
	}
}

// SetStreamHandler registers a libp2p protocol stream handler
func (n *Node) SetStreamHandler(proto string, handler func(network.Stream)) {
	n.Host.SetStreamHandler(protocol.ID(proto), handler)
}

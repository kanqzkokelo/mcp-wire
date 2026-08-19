package p2p

import (
	"context"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const ServiceTag = "mcp-wire-mdns"

type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return
	}
	slog.Debug("mDNS discovered LAN peer", "peer", pi.ID.String())
	n.h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.AddressTTL)
}

// InitMDNS initializes mDNS service for zero-config local network peer discovery
func InitMDNS(ctx context.Context, h host.Host) (mdns.Service, error) {
	notifee := &discoveryNotifee{h: h}
	svc := mdns.NewMdnsService(h, ServiceTag, notifee)
	if err := svc.Start(); err != nil {
		return nil, err
	}
	slog.Info("mDNS local peer discovery service started")
	return svc, nil
}

package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func DialStream(ctx context.Context, h host.Host, targetInfo peer.AddrInfo, protoID protocol.ID) (network.Stream, error) {
	if len(targetInfo.Addrs) > 0 {
		h.Peerstore().AddAddrs(targetInfo.ID, targetInfo.Addrs, peerstore.PermanentAddrTTL)
	}

	// Connect to peer if not connected
	if h.Network().Connectedness(targetInfo.ID) != network.Connected {
		dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		if err := h.Connect(dialCtx, targetInfo); err != nil {
			return nil, fmt.Errorf("failed to connect to peer %s: %w", targetInfo.ID.String(), err)
		}
	}

	// Open stream
	stream, err := h.NewStream(ctx, targetInfo.ID, protoID)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream %s to peer %s: %w", string(protoID), targetInfo.ID.String(), err)
	}

	return stream, nil
}

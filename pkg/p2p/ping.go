package p2p

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

func PingPeer(ctx context.Context, h host.Host, target peer.ID) time.Duration {
	pService := ping.NewPingService(h)
	res := pService.Ping(ctx, target)
	select {
	case <-ctx.Done():
		return 0
	case r := <-res:
		if r.Error != nil {
			return 0
		}
		return r.RTT
	}
}

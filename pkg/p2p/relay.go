package p2p

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	circuitv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
)

func ReserveRelay(ctx context.Context, h host.Host, relayInfo peer.AddrInfo) (*circuitv2.Reservation, error) {
	return circuitv2.Reserve(ctx, h, relayInfo)
}

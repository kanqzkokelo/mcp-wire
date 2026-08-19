package p2p

import (
	"context"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
)

func ListenEvents(ctx context.Context, h host.Host) {
	sub, err := h.EventBus().Subscribe(new(event.EvtPeerConnectednessChanged))
	if err != nil {
		slog.Debug("failed to subscribe to peer connectedness events", "error", err)
		return
	}

	go func() {
		defer sub.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-sub.Out():
				if !ok {
					return
				}
				if e, ok := evt.(event.EvtPeerConnectednessChanged); ok {
					slog.Debug("Peer connectedness changed", "peer", e.Peer.String(), "state", e.Connectedness.String())
				}
			}
		}
	}()
}

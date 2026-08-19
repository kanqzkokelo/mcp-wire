package cmd

import (
	"fmt"

	"github.com/mcp-wire/mcp-wire/pkg/identity"
	"github.com/mcp-wire/mcp-wire/pkg/metrics"
	"github.com/mcp-wire/mcp-wire/pkg/p2p"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display local mcp-wire daemon status, peer ID, and diagnostic metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		privKey, err := identity.LoadOrGenerateIdentity()
		if err != nil {
			return fmt.Errorf("identity check failed: %w", err)
		}

		pid, _ := identity.GetPeerID(privKey)

		snapshot := metrics.GetStatusSnapshot(
			pid.String(),
			p2p.DefaultListenAddrs(0),
			[]string{},
			map[string]string{},
		)

		if jsonOut {
			jsonStr, err := metrics.ExportStatusJSON(snapshot)
			if err != nil {
				return err
			}
			fmt.Println(jsonStr)
			return nil
		}

		fmt.Println("=== mcp-wire Status & Diagnostics ===")
		fmt.Printf("Peer ID:           %s\n", snapshot.PeerID)
		fmt.Printf("Active Streams:    %d\n", snapshot.ActiveStreams)
		fmt.Printf("Total Connections: %d\n", snapshot.TotalConnections)
		fmt.Printf("Bytes Sent:        %d B\n", snapshot.BytesSent)
		fmt.Printf("Bytes Received:    %d B\n", snapshot.BytesReceived)
		fmt.Printf("Heap Alloc:        %d B\n", snapshot.HeapAllocBytes)
		fmt.Printf("Goroutines:        %d\n", snapshot.NumGoroutines)
		fmt.Printf("Uptime:            %s\n", snapshot.Uptime)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

package cmd

import (
	"fmt"
	"os"

	"github.com/mcp-wire/mcp-wire/pkg/auth"
	"github.com/mcp-wire/mcp-wire/pkg/config"
	"github.com/mcp-wire/mcp-wire/pkg/identity"
	"github.com/mcp-wire/mcp-wire/pkg/p2p"
	"github.com/mcp-wire/mcp-wire/pkg/protocol"
	"github.com/mcp-wire/mcp-wire/pkg/proxy"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
)

var (
	connectToken string
	connectAddr  string
)

var connectCmd = &cobra.Command{
	Use:   "connect <mcp://PeerID/ServiceName>",
	Short: "Bridge remote MCP server over P2P wire to standard stdio for local AI agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURI := args[0]
		mcpURI, err := protocol.ParseURI(rawURI)
		if err != nil {
			return err
		}

		token := connectToken
		if token == "" {
			token = mcpURI.Token
		}

		ctx, cancel := config.SetupSignalContext()
		defer cancel()

		privKey, err := identity.LoadOrGenerateIdentity()
		if err != nil {
			return fmt.Errorf("identity setup failed: %w", err)
		}

		node, err := p2p.NewHost(ctx, privKey, p2p.DefaultListenAddrs(0))
		if err != nil {
			return err
		}
		defer node.Close()

		dhtSvc, err := p2p.InitDHT(ctx, node.Host)
		if err == nil {
			go dhtSvc.ConnectBootstrapNodes(ctx, nil)
		}

		protoID := protocol.ToProtocolID(mcpURI.ServiceName)

		// Dial remote peer
		targetInfo := node.Host.Peerstore().PeerInfo(mcpURI.PeerID)
		if connectAddr != "" {
			maddr, err := ma.NewMultiaddr(connectAddr)
			if err == nil {
				targetInfo.Addrs = append(targetInfo.Addrs, maddr)
			}
		}
		if len(targetInfo.Addrs) == 0 {
			// Lookup addrs via DHT
			if dhtSvc != nil {
				foundInfo, err := dhtSvc.FindPeerAddrs(ctx, mcpURI.PeerID)
				if err == nil {
					targetInfo = foundInfo
				}
			}
		}

		stream, err := p2p.DialStream(ctx, node.Host, targetInfo, protoID)
		if err != nil {
			return fmt.Errorf("failed to connect to remote MCP server: %w", err)
		}
		defer stream.Close()

		// Perform Client Auth Handshake
		if err := auth.PerformClientHandshake(stream, token, mcpURI.ServiceName); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Bridge local os.Stdin & os.Stdout to P2P stream transparently
		proxy.BridgeStdioToStream(ctx, stream, os.Stdin, os.Stdout)
		return nil
	},
}

func init() {
	connectCmd.Flags().StringVar(&connectToken, "token", "", "pre-shared secret authorization token")
	connectCmd.Flags().StringVar(&connectAddr, "addr", "", "explicit multiaddr target hint (e.g. /ip4/127.0.0.1/tcp/49152)")
	rootCmd.AddCommand(connectCmd)
}

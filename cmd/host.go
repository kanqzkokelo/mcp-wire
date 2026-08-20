package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/mcp-wire/mcp-wire/pkg/auth"
	"github.com/mcp-wire/mcp-wire/pkg/config"
	"github.com/mcp-wire/mcp-wire/pkg/identity"
	"github.com/mcp-wire/mcp-wire/pkg/metrics"
	"github.com/mcp-wire/mcp-wire/pkg/p2p"
	"github.com/mcp-wire/mcp-wire/pkg/protocol"
	"github.com/mcp-wire/mcp-wire/pkg/proxy"
	"github.com/spf13/cobra"
)

var (
	hostName             string
	hostCmdStr           string
	allowPeerIDs         []string
	authToken            string
	allowUnauthenticated bool
	readOnlyMode         bool
	listenPort           int
	rateLimit            float64
	rateBurst            int
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Share a local MCP server over the P2P mesh wire",
	RunE: func(cmd *cobra.Command, args []string) error {
		if hostName == "" {
			return fmt.Errorf("--name is required (e.g. --name gpu-whisper)")
		}
		if hostCmdStr == "" {
			return fmt.Errorf("--cmd is required (e.g. --cmd \"python server.py\")")
		}

		if err := protocol.ValidateServiceName(hostName); err != nil {
			return err
		}

		ctx, cancel := config.SetupSignalContext()
		defer cancel()

		// Load or generate identity key
		privKey, err := identity.LoadOrGenerateIdentity()
		if err != nil {
			return fmt.Errorf("identity setup failed: %w", err)
		}

		peerID, _ := identity.GetPeerID(privKey)

		// Setup ACL
		acl, err := auth.LoadACL("")
		if err != nil {
			return fmt.Errorf("ACL setup failed: %w", err)
		}

		for _, allowed := range allowPeerIDs {
			_ = acl.AddAllowedPeer(allowed)
		}

		// Fallback token from environment variable if unset
		if authToken == "" {
			authToken = os.Getenv("MCP_WIRE_TOKEN")
		}

		// Secure default: enforce authentication unless unauthenticated mode explicitly allowed
		if authToken == "" && len(allowPeerIDs) == 0 && len(acl.AllowedPeers) == 0 {
			if !allowUnauthenticated {
				autoToken, err := auth.GenerateSecureToken(16)
				if err != nil {
					return fmt.Errorf("failed to generate secure auth token: %w", err)
				}
				authToken = autoToken
				slog.Info("No auth token or ACL specified. Auto-generated secret auth token for host.")
			} else {
				slog.Warn("SECURITY WARNING: Host running in unauthenticated mode (--allow-unauthenticated). Any P2P node can connect!")
			}
		}

		// Create libp2p host
		node, err := p2p.NewHost(ctx, privKey, p2p.DefaultListenAddrs(listenPort))
		if err != nil {
			return err
		}
		defer node.Close()

		// Init DHT, AutoNAT & mDNS
		dhtSvc, err := p2p.InitDHT(ctx, node.Host)
		if err == nil {
			go dhtSvc.ConnectBootstrapNodes(ctx, nil)
		}
		_, _ = p2p.InitMDNS(ctx, node.Host)

		protoID := protocol.ToProtocolID(hostName)

		var peerLimiters sync.Map

		// Register P2P stream handler
		node.Host.SetStreamHandler(protoID, func(s network.Stream) {
			metrics.StreamOpened()
			defer metrics.StreamClosed()
			defer s.Close()

			// Rate limiting per peer
			remotePeer := s.Conn().RemotePeer()
			limVal, _ := peerLimiters.LoadOrStore(remotePeer, auth.NewRateLimiter(rateLimit, rateBurst))
			if !limVal.(*auth.RateLimiter).Allow() {
				slog.Warn("rate limit exceeded for remote peer", "peer", remotePeer.String())
				return
			}

			// Perform Auth Handshake
			hsFrame, err := auth.PerformHostHandshake(s, authToken, acl)
			if err != nil {
				slog.Warn("host handshake failed", "peer", remotePeer.String(), "error", err)
				return
			}

			slog.Info("client authenticated successfully", "peer", s.Conn().RemotePeer().String(), "service", hsFrame.Service)

			// Launch target MCP server subprocess
			tp, err := proxy.SpawnTargetProcess(ctx, hostCmdStr, s.Conn().RemotePeer().String())
			if err != nil {
				slog.Error("failed to spawn target process", "error", err)
				return
			}
			defer tp.Stop()

			// Setup stream readers & writers with metrics and read-only filter
			var processIn io.WriteCloser = tp.Stdin()
			var processOut io.ReadCloser = tp.Stdout()

			if readOnlyMode {
				filter := auth.NewReadOnlyFilterReader(s, true)
				// Process stdout -> P2P stream
				go func() {
					buf := make([]byte, 32*1024)
					for {
						n, rerr := processOut.Read(buf)
						if n > 0 {
							_, _ = s.Write(buf[:n])
						}
						if rerr != nil {
							break
						}
					}
				}()

				// Stream -> filter -> allowed to processIn, blocked back to s
				for {
					if err := filter.ReadLineAndFilter(processIn, s); err != nil {
						break
					}
				}
				_ = processIn.Close()
				_ = tp.Stop()
				return
			}

			proxy.BridgeStreams(ctx, s, processIn, processOut)
		})

		connURI := protocol.BuildURI(peerID, hostName, authToken)

		fmt.Printf("\n🚀 mcp-wire daemon running!\n")
		fmt.Printf("   Peer ID: %s\n", peerID.String())
		fmt.Printf("   Service: %s\n", string(protoID))
		fmt.Printf("   Connection String: %s\n\n", connURI)
		fmt.Printf("Paste this into your client's mcp-wire connect command.\n\n")

		<-ctx.Done()
		fmt.Println("\nShutting down host daemon...")
		return nil
	},
}

func init() {
	hostCmd.Flags().StringVar(&hostName, "name", "", "name of the hosted MCP service (e.g. gpu-whisper)")
	hostCmd.Flags().StringVar(&hostCmdStr, "cmd", "", "command to execute target MCP server (e.g. \"python server.py\")")
	hostCmd.Flags().StringSliceVar(&allowPeerIDs, "allow", nil, "allowed client peer IDs")
	hostCmd.Flags().StringVar(&authToken, "token", "", "optional pre-shared authorization token secret (or set via MCP_WIRE_TOKEN env)")
	hostCmd.Flags().BoolVar(&allowUnauthenticated, "allow-unauthenticated", false, "explicitly allow unauthenticated client connections (insecure)")
	hostCmd.Flags().BoolVar(&readOnlyMode, "read-only", false, "enforce read-only mode blocking mutating tool calls")
	hostCmd.Flags().IntVar(&listenPort, "port", 0, "p2p listening port (0 for random)")
	hostCmd.Flags().Float64Var(&rateLimit, "rate-limit", 100.0, "maximum incoming requests per second per peer")
	hostCmd.Flags().IntVar(&rateBurst, "rate-burst", 200, "maximum burst requests per peer")
	rootCmd.AddCommand(hostCmd)
}

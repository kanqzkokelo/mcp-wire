package cmd

import (
	"fmt"
	"os"

	"github.com/mcp-wire/mcp-wire/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	verbose  bool
	jsonOut  bool
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "mcp-wire",
	Short: "Zero-Config P2P Mesh for Model Context Protocol (MCP)",
	Long: `mcp-wire is a single-binary daemon that enables zero-config, peer-to-peer (P2P)
distribution of Model Context Protocol (MCP) tools across local and remote devices
using libp2p, Ed25519 identity, and E2E Noise transport security.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cfgFile != "" {
			config.CustomConfigDir = cfgFile
		}
		config.InitLogger(logLevel, verbose)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config directory (default is ~/.mcp-wire)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose debug logging")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "format output as JSON")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
}

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/mcp-wire/mcp-wire/pkg/protocol"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config <mcp://PeerID/ServiceName>",
	Short: "Generate copy-paste JSON snippet for Claude Desktop / OpenCode configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURI := args[0]
		mcpURI, err := protocol.ParseURI(rawURI)
		if err != nil {
			return err
		}

		type MCPServerConfig struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}

		type MCPConfigWrapper struct {
			MCPServers map[string]MCPServerConfig `json:"mcpServers"`
		}

		connArgs := []string{"connect", mcpURI.Raw}
		if mcpURI.Token != "" {
			connArgs = append(connArgs, "--token", mcpURI.Token)
		}

		cfg := MCPConfigWrapper{
			MCPServers: map[string]MCPServerConfig{
				mcpURI.ServiceName: {
					Command: "mcp-wire",
					Args:    connArgs,
				},
			},
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println("=== Add this snippet to your claude_desktop_config.json / opencode.jsonc ===")
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

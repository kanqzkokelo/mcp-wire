package cmd

import (
	"fmt"

	"github.com/mcp-wire/mcp-wire/pkg/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage access control list (allowed peers)",
}

var authAddCmd = &cobra.Command{
	Use:   "add <peer-id>",
	Short: "Add a peer ID to allowed peers whitelist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		peerIDStr := args[0]
		acl, err := auth.LoadACL("")
		if err != nil {
			return err
		}
		if err := acl.AddAllowedPeer(peerIDStr); err != nil {
			return err
		}
		if err := auth.SaveACL("", acl); err != nil {
			return err
		}
		fmt.Printf("Added peer '%s' to allowed peers list.\n", peerIDStr)
		return nil
	},
}

var authRemoveCmd = &cobra.Command{
	Use:   "remove <peer-id>",
	Short: "Remove a peer ID from allowed peers whitelist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		peerIDStr := args[0]
		acl, err := auth.LoadACL("")
		if err != nil {
			return err
		}
		if acl.RemoveAllowedPeer(peerIDStr) {
			_ = auth.SaveACL("", acl)
			fmt.Printf("Removed peer '%s' from allowed peers list.\n", peerIDStr)
		} else {
			fmt.Printf("Peer '%s' not found in allowed peers list.\n", peerIDStr)
		}
		return nil
	},
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all allowed peer IDs",
	RunE: func(cmd *cobra.Command, args []string) error {
		acl, err := auth.LoadACL("")
		if err != nil {
			return err
		}
		peers := acl.ListAllowedPeers()
		if len(peers) == 0 {
			fmt.Println("Allowed Peers: [empty] (Default: accepting connections)")
			return nil
		}
		fmt.Println("Allowed Peers:")
		for _, p := range peers {
			fmt.Printf(" - %s\n", p)
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authAddCmd)
	authCmd.AddCommand(authRemoveCmd)
	authCmd.AddCommand(authListCmd)
	rootCmd.AddCommand(authCmd)
}

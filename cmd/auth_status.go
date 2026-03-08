package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long:  `Display the current authentication status and configuration.`,
	RunE:  runAuthStatus,
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()
	manager := getAuthManager()

	status := manager.GetAuthStatus()
	output.PrintOutput(status, textMode, "")

	return nil
}

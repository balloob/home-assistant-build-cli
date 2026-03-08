package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long:  `Remove stored credentials from the local machine.`,
	RunE:  runAuthLogout,
}

func init() {
	authCmd.AddCommand(authLogoutCmd)
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()
	manager := getAuthManager()

	if err := manager.Logout(); err != nil {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	output.PrintSuccess(nil, textMode, "Successfully logged out.")
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")
	manager := auth.NewManager(configDir)

	status := manager.GetAuthStatus()
	output.PrintOutput(status, textMode, "")

	return nil
}

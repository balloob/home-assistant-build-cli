package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Force token refresh (OAuth only)",
	Long:  `Force a refresh of the OAuth access token.`,
	RunE:  runAuthRefresh,
}

func init() {
	authCmd.AddCommand(authRefreshCmd)
}

func runAuthRefresh(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")
	manager := auth.NewManager(configDir)

	creds, err := manager.GetCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}
	if creds == nil {
		return fmt.Errorf("not authenticated")
	}

	if !creds.IsOAuth() {
		return fmt.Errorf("token refresh is only available for OAuth authentication")
	}

	if err := manager.RefreshToken(); err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	creds, _ = manager.GetCredentials()
	result := map[string]interface{}{
		"token_expiry": creds.TokenExpiry,
	}

	output.PrintSuccess(result, textMode, "Token refreshed successfully.")
	return nil
}

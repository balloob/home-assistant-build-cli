package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long: `Remove the selected context's stored credentials from the local machine.

Other contexts are not affected. Logging out of a named context removes the
context; logging out of the default context removes only its credentials.`,
	Example: `  hab auth logout
  hab auth logout --context brother-homeassistant`,
	RunE: runAuthLogout,
}

func init() {
	authCmd.AddCommand(authLogoutCmd)
	mergeSchemaAnnotation(authLogoutCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"local"},
		ResourceType: "auth",
		GuideTopic:   "auth",
	})
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

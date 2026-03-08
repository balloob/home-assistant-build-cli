package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	textMode := viper.GetBool("text")
	manager := getAuthManager()

	if manager.Logout() {
		output.PrintSuccess(nil, textMode, "Successfully logged out.")
	} else {
		output.PrintSuccess(nil, textMode, "No credentials to remove.")
	}

	return nil
}

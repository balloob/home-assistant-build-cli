package cmd

import (
	"github.com/spf13/cobra"
)

var notificationCmd = &cobra.Command{
	Use:   "notification",
	Short: "Manage persistent notifications",
	Long: `List, create, and dismiss persistent notifications in Home Assistant.

For workflow guidance, run 'hab guide operations'.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(notificationCmd)
}

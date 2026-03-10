package cmd

import (
	"github.com/spf13/cobra"
)

var notificationCmd = &cobra.Command{
	Use:     "notification",
	Short:   "Manage persistent notifications",
	Long:    `List, create, and dismiss persistent notifications in Home Assistant.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(notificationCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
)

var badgeCmd = &cobra.Command{
	Use:   "badge",
	Short: "Manage view badges",
	Long:  `Create, update, list, and delete badges in a dashboard view.`,
}

func init() {
	rootCmd.AddCommand(badgeCmd)
}

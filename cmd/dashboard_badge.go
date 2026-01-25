package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardBadgeCmd = &cobra.Command{
	Use:   "dashboard-badge",
	Short: "Manage view badges",
	Long:  `Create, update, list, and delete badges in a dashboard view.`,
}

func init() {
	rootCmd.AddCommand(dashboardBadgeCmd)
}

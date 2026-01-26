package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardBadgeCmd = &cobra.Command{
	Use:     "badge",
	Short:   "Manage view badges",
	Long:    `Create, update, list, and delete badges in a dashboard view.`,
	GroupID: dashboardGroupSubcommands,
}

func init() {
	dashboardCmd.AddCommand(dashboardBadgeCmd)
}

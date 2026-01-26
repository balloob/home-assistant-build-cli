package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardCardCmd = &cobra.Command{
	Use:     "card",
	Short:   "Manage dashboard cards",
	Long:    `Create, update, list, and delete cards in a dashboard view or section.`,
	GroupID: dashboardGroupSubcommands,
}

func init() {
	dashboardCmd.AddCommand(dashboardCardCmd)
}

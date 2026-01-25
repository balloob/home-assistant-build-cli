package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardCardCmd = &cobra.Command{
	Use:   "dashboard-card",
	Short: "Manage dashboard cards",
	Long:  `Create, update, list, and delete cards in a dashboard view or section.`,
}

func init() {
	rootCmd.AddCommand(dashboardCardCmd)
}

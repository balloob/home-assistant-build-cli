package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardViewCmd = &cobra.Command{
	Use:     "view",
	Short:   "Manage dashboard views",
	Long:    `Create, update, list, and delete views in a dashboard.`,
	GroupID: dashboardGroupSubcommands,
}

func init() {
	dashboardCmd.AddCommand(dashboardViewCmd)
}

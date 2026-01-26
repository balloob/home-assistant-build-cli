package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardSectionCmd = &cobra.Command{
	Use:     "section",
	Short:   "Manage view sections",
	Long:    `Create, update, list, and delete sections in a dashboard view.`,
	GroupID: dashboardGroupSubcommands,
}

func init() {
	dashboardCmd.AddCommand(dashboardSectionCmd)
}

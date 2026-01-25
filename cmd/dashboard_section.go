package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardSectionCmd = &cobra.Command{
	Use:   "dashboard-section",
	Short: "Manage view sections",
	Long:  `Create, update, list, and delete sections in a dashboard view.`,
}

func init() {
	rootCmd.AddCommand(dashboardSectionCmd)
}

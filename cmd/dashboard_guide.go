package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardGuideCmd = &cobra.Command{
	Use:     "guide",
	Short:   "Display the dashboard creation guide",
	Long:    `Display best practices and tips for creating effective Home Assistant dashboards.`,
	GroupID: dashboardGroupCommands,
	RunE:    runDashboardGuide,
}

func init() {
	dashboardCmd.AddCommand(dashboardGuideCmd)
}

func runDashboardGuide(cmd *cobra.Command, args []string) error {
	return runGuideTopic("dashboard")
}

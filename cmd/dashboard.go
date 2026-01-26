package cmd

import (
	"github.com/spf13/cobra"
)

const (
	dashboardGroupCommands    = "commands"
	dashboardGroupSubcommands = "subcommands"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Manage dashboards",
	Long:  `Create, update, and delete dashboards.`,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)

	dashboardCmd.AddGroup(
		&cobra.Group{ID: dashboardGroupCommands, Title: "Commands:"},
		&cobra.Group{ID: dashboardGroupSubcommands, Title: "Subcommands:"},
	)
}

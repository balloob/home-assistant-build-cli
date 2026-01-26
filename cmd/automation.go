package cmd

import (
	"github.com/spf13/cobra"
)

const (
	automationGroupCommands    = "commands"
	automationGroupSubcommands = "subcommands"
)

var automationCmd = &cobra.Command{
	Use:   "automation",
	Short: "Manage automations",
	Long:  `Create, update, delete, and trigger automations.`,
}

func init() {
	rootCmd.AddCommand(automationCmd)

	automationCmd.AddGroup(
		&cobra.Group{ID: automationGroupCommands, Title: "Commands:"},
		&cobra.Group{ID: automationGroupSubcommands, Title: "Subcommands:"},
	)
}

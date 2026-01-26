package cmd

import (
	"github.com/spf13/cobra"
)

const (
	helperGroupCommands    = "commands"
	helperGroupSubcommands = "subcommands"
)

var helperCmd = &cobra.Command{
	Use:     "helper",
	Short:   "Manage groups, templates, and other helpers",
	Long:    `Create, update, and delete helper entities like input booleans, counters, timers, groups, templates, and more.`,
	GroupID: "automation",
}

func init() {
	rootCmd.AddCommand(helperCmd)

	helperCmd.AddGroup(
		&cobra.Group{ID: helperGroupCommands, Title: "Commands:"},
		&cobra.Group{ID: helperGroupSubcommands, Title: "Helper Types:"},
	)
}

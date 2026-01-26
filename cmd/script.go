package cmd

import (
	"github.com/spf13/cobra"
)

const (
	scriptGroupCommands    = "commands"
	scriptGroupSubcommands = "subcommands"
)

var scriptCmd = &cobra.Command{
	Use:     "script",
	Short:   "Manage scripts",
	Long:    `Create, update, delete, and run scripts.`,
	GroupID: "automation",
}

func init() {
	rootCmd.AddCommand(scriptCmd)

	scriptCmd.AddGroup(
		&cobra.Group{ID: scriptGroupCommands, Title: "Commands:"},
		&cobra.Group{ID: scriptGroupSubcommands, Title: "Subcommands:"},
	)
}

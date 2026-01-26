package cmd

import (
	"github.com/spf13/cobra"
)

var helperLocalCalendarParentCmd = &cobra.Command{
	Use:     "local-calendar",
	Short:   "Manage local calendar helpers",
	Long:    `Create, list, and delete local calendar helpers for storing calendar events locally.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperLocalCalendarParentCmd)
}

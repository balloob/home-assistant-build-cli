package cmd

import (
	"github.com/spf13/cobra"
)

var helperScheduleParentCmd = &cobra.Command{
	Use:     "schedule",
	Short:   "Manage schedule helpers",
	Long:    `Create, list, and delete schedule helpers for time-based automation.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperScheduleParentCmd)
}

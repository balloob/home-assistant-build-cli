package cmd

import (
	"github.com/spf13/cobra"
)

var helperThresholdParentCmd = &cobra.Command{
	Use:     "threshold",
	Short:   "Manage threshold binary sensor helpers",
	Long:    `Create, list, and delete threshold binary sensor helpers that monitor sensor values against thresholds.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperThresholdParentCmd)
}

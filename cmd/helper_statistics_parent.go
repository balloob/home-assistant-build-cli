package cmd

import (
	"github.com/spf13/cobra"
)

var helperStatisticsParentCmd = &cobra.Command{
	Use:     "statistics",
	Short:   "Manage statistics sensor helpers",
	Long:    `Create, list, and delete statistics sensor helpers that provide statistical analysis of sensor history.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperStatisticsParentCmd)
}

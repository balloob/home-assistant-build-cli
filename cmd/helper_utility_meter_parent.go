package cmd

import (
	"github.com/spf13/cobra"
)

var helperUtilityMeterParentCmd = &cobra.Command{
	Use:     "utility-meter",
	Short:   "Manage utility meter helpers",
	Long:    `Create, list, and delete utility meter helpers that track consumption across billing cycles.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperUtilityMeterParentCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
)

var helperUtilityMeterParentCmd = &cobra.Command{
	Use:   "helper-utility-meter",
	Short: "Manage utility meter helpers",
	Long:  `Create, list, and delete utility meter helpers that track consumption across billing cycles.`,
}

func init() {
	rootCmd.AddCommand(helperUtilityMeterParentCmd)
}

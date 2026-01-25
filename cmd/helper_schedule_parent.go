package cmd

import (
	"github.com/spf13/cobra"
)

var helperScheduleParentCmd = &cobra.Command{
	Use:   "helper-schedule",
	Short: "Manage schedule helpers",
	Long:  `Create, list, and delete schedule helpers for time-based automation.`,
}

func init() {
	rootCmd.AddCommand(helperScheduleParentCmd)
}

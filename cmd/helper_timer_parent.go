package cmd

import (
	"github.com/spf13/cobra"
)

var helperTimerParentCmd = &cobra.Command{
	Use:   "helper-timer",
	Short: "Manage timer helpers",
	Long:  `Create, list, and delete timer helpers.`,
}

func init() {
	rootCmd.AddCommand(helperTimerParentCmd)
}

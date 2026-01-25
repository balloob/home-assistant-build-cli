package cmd

import (
	"github.com/spf13/cobra"
)

var automationActionCmd = &cobra.Command{
	Use:   "automation-action",
	Short: "Manage automation actions",
	Long:  `Create, update, list, and delete actions in an automation.`,
}

func init() {
	rootCmd.AddCommand(automationActionCmd)
}

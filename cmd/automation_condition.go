package cmd

import (
	"github.com/spf13/cobra"
)

var automationConditionCmd = &cobra.Command{
	Use:   "automation-condition",
	Short: "Manage automation conditions",
	Long:  `Create, update, list, and delete conditions in an automation.`,
}

func init() {
	rootCmd.AddCommand(automationConditionCmd)
}

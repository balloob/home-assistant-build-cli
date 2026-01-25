package cmd

import (
	"github.com/spf13/cobra"
)

var automationTriggerParentCmd = &cobra.Command{
	Use:   "automation-trigger",
	Short: "Manage automation triggers",
	Long:  `Create, update, list, and delete triggers in an automation.`,
}

func init() {
	rootCmd.AddCommand(automationTriggerParentCmd)
}

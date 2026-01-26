package cmd

import (
	"github.com/spf13/cobra"
)

var automationConditionCmd = &cobra.Command{
	Use:     "condition",
	Short:   "Manage automation conditions",
	Long:    `Create, update, list, and delete conditions in an automation.`,
	GroupID: automationGroupSubcommands,
}

func init() {
	automationCmd.AddCommand(automationConditionCmd)
}

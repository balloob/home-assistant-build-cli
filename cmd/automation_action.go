package cmd

import (
	"github.com/spf13/cobra"
)

var automationActionCmd = &cobra.Command{
	Use:     "action",
	Short:   "Manage automation actions",
	Long:    `Create, update, list, and delete actions in an automation.`,
	GroupID: automationGroupSubcommands,
}

func init() {
	automationCmd.AddCommand(automationActionCmd)
}

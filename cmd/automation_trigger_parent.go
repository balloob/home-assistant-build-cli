package cmd

import (
	"github.com/spf13/cobra"
)

var automationTriggerParentCmd = &cobra.Command{
	Use:     "trigger",
	Short:   "Manage automation triggers",
	Long:    `Create, update, list, and delete triggers in an automation.`,
	GroupID: automationGroupSubcommands,
}

func init() {
	automationCmd.AddCommand(automationTriggerParentCmd)
}

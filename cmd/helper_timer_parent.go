package cmd

import (
	"github.com/spf13/cobra"
)

var helperTimerParentCmd = &cobra.Command{
	Use:     "timer",
	Short:   "Manage timer helpers",
	Long:    `Create, list, and delete timer helpers.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperTimerParentCmd)
}

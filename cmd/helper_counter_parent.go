package cmd

import (
	"github.com/spf13/cobra"
)

var helperCounterParentCmd = &cobra.Command{
	Use:     "counter",
	Short:   "Manage counter helpers",
	Long:    `Create, list, and delete counter helpers.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperCounterParentCmd)
}

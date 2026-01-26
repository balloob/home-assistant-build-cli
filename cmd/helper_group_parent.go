package cmd

import (
	"github.com/spf13/cobra"
)

var helperGroupParentCmd = &cobra.Command{
	Use:     "group",
	Short:   "Manage group helpers",
	Long:    `Create, list, and delete group helpers.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperGroupParentCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputSelectParentCmd = &cobra.Command{
	Use:     "input-select",
	Short:   "Manage input select helpers",
	Long:    `Create, list, and delete input select (dropdown) helpers.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperInputSelectParentCmd)
}

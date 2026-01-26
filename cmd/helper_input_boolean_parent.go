package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputBooleanParentCmd = &cobra.Command{
	Use:     "input-boolean",
	Short:   "Manage input boolean helpers",
	Long:    `Create, list, and delete input boolean (toggle) helpers.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperInputBooleanParentCmd)
}

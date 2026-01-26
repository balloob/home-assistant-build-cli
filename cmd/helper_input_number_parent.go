package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputNumberParentCmd = &cobra.Command{
	Use:     "input-number",
	Short:   "Manage input number helpers",
	Long:    `Create, list, and delete input number helpers.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperInputNumberParentCmd)
}

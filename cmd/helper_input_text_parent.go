package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputTextParentCmd = &cobra.Command{
	Use:     "input-text",
	Short:   "Manage input text helpers",
	Long:    `Create, list, and delete input text helpers.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperInputTextParentCmd)
}

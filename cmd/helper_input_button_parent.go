package cmd

import (
	"github.com/spf13/cobra"
)

var helperInputButtonParentCmd = &cobra.Command{
	Use:     "input-button",
	Short:   "Manage input button helpers",
	Long:    `Create, list, and delete input button helpers.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperInputButtonParentCmd)
}

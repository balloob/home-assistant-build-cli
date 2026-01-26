package cmd

import (
	"github.com/spf13/cobra"
)

var helperLocalTodoParentCmd = &cobra.Command{
	Use:     "local-todo",
	Short:   "Manage local to-do list helpers",
	Long:    `Create, list, and delete local to-do list helpers for storing tasks locally.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperLocalTodoParentCmd)
}

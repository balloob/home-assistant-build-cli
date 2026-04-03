package cmd

import (
	"github.com/spf13/cobra"
)

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "Manage to-do list items",
	Long: `List, add, complete, and remove items from Home Assistant to-do lists.

For workflow guidance, run 'hab guide calendar-todo'.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(todoCmd)
}

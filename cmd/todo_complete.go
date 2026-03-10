package cmd

import (
	"github.com/spf13/cobra"
)

var todoCompleteCmd = &cobra.Command{
	Use:   "complete <entity_id> <item_uid>",
	Short: "Mark a to-do item as complete",
	Long:  `Mark a to-do list item as completed.`,
	Example: `  hab todo complete todo.shopping_list abc123
  hab todo complete shopping_list abc123`,
	Args: cobra.ExactArgs(2),
	RunE: runTodoComplete,
}

func init() {
	todoCmd.AddCommand(todoCompleteCmd)
}

func runTodoComplete(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "todo")
	uid := args[1]

	return callServiceAction("todo", "update_item", "Item marked complete.", map[string]interface{}{
		"entity_id": entityID,
		"item":      uid,
		"status":    "completed",
	})
}

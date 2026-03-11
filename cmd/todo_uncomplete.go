package cmd

import (
	"github.com/spf13/cobra"
)

var todoUncompleteCmd = &cobra.Command{
	Use:   "uncomplete <entity_id> <item_uid>",
	Short: "Mark a to-do item as not complete",
	Long:  `Mark a to-do list item as needing action (undo completion).`,
	Example: `  hab todo uncomplete todo.shopping_list abc123`,
	Args: cobra.ExactArgs(2),
	RunE: runTodoUncomplete,
}

func init() {
	todoCmd.AddCommand(todoUncompleteCmd)
}

func runTodoUncomplete(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "todo")
	uid := args[1]

	return callServiceAction("todo", "update_item", "Item marked incomplete.", map[string]interface{}{
		"entity_id": entityID,
		"item":      uid,
		"status":    "needs_action",
	})
}

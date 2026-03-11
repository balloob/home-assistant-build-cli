package cmd

import (
	"github.com/spf13/cobra"
)

var todoRemoveCmd = &cobra.Command{
	Use:   "remove <entity_id> <item_uid>",
	Short: "Remove an item from a to-do list",
	Long:  `Remove an item from a Home Assistant to-do list by its uid.`,
	Example: `  hab todo remove todo.shopping_list abc123`,
	Args: cobra.ExactArgs(2),
	RunE: runTodoRemove,
}

func init() {
	todoCmd.AddCommand(todoRemoveCmd)
}

func runTodoRemove(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "todo")
	uid := args[1]

	return callServiceAction("todo", "remove_item", "Item removed.", map[string]interface{}{
		"entity_id": entityID,
		"item":      uid,
	})
}

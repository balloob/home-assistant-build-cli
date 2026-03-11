package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var todoItemsCmd = &cobra.Command{
	Use:   "items <entity_id>",
	Short: "List items in a to-do list",
	Long:  `List all items in a to-do list entity.`,
	Example: `  hab todo items todo.shopping_list
  hab todo items todo.shopping_list --json`,
	Args: cobra.ExactArgs(1),
	RunE: runTodoItems,
}

func init() {
	todoCmd.AddCommand(todoItemsCmd)
}

func runTodoItems(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "todo")
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	items, err := ws.TodoItemList(entityID)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		output.PrintOutput([]interface{}{}, textMode, "No items found.")
		return nil
	}

	output.PrintOutput(items, textMode, "")
	return nil
}

package cmd

import (
	"github.com/spf13/cobra"
)

var (
	todoAddDue         string
	todoAddDescription string
)

var todoAddCmd = &cobra.Command{
	Use:   "add <entity_id> <summary>",
	Short: "Add an item to a to-do list",
	Long:  `Add a new item to a Home Assistant to-do list.`,
	Example: `  hab todo add todo.shopping_list "Milk"
  hab todo add todo.shopping_list "Doctor appointment" --due 2026-04-01
  hab todo add todo.shopping_list "Call plumber" --due 2026-04-01T09:00:00 --description "Re: kitchen leak"`,
	Args: cobra.ExactArgs(2),
	RunE: runTodoAdd,
}

func init() {
	todoCmd.AddCommand(todoAddCmd)
	todoAddCmd.Flags().StringVar(&todoAddDue, "due", "", "Due date or datetime (ISO 8601: YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)")
	todoAddCmd.Flags().StringVar(&todoAddDescription, "description", "", "Optional description for the item")
}

func runTodoAdd(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "todo")
	summary := args[1]

	data := map[string]interface{}{
		"entity_id": entityID,
		"item":      summary,
	}
	if todoAddDescription != "" {
		data["description"] = todoAddDescription
	}
	// HA distinguishes due_date (YYYY-MM-DD) from due_datetime (with time component)
	if todoAddDue != "" {
		if len(todoAddDue) > 10 {
			data["due_datetime"] = todoAddDue
		} else {
			data["due_date"] = todoAddDue
		}
	}

	return callServiceAction("todo", "add_item", "Item added.", data)
}

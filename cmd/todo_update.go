package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	todoUpdateSummary     string
	todoUpdateDue         string
	todoUpdateDescription string
	todoUpdateStatus      string
)

var todoUpdateCmd = &cobra.Command{
	Use:   "update <entity_id> <item_uid>",
	Short: "Update a to-do item",
	Long:  `Update the summary, due date, description, or status of a to-do list item.`,
	Example: `  hab todo update todo.shopping_list abc123 --summary "Oat milk"
  hab todo update todo.shopping_list abc123 --due 2026-04-15
  hab todo update todo.shopping_list abc123 --status completed`,
	Args: cobra.ExactArgs(2),
	RunE: runTodoUpdate,
}

func init() {
	todoCmd.AddCommand(todoUpdateCmd)
	todoUpdateCmd.Flags().StringVar(&todoUpdateSummary, "summary", "", "New summary/title for the item")
	todoUpdateCmd.Flags().StringVar(&todoUpdateDue, "due", "", "New due date or datetime (ISO 8601)")
	todoUpdateCmd.Flags().StringVar(&todoUpdateDescription, "description", "", "New description for the item")
	todoUpdateCmd.Flags().StringVar(&todoUpdateStatus, "status", "", "New status: needs_action or completed")
}

func runTodoUpdate(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "todo")
	uid := args[1]

	if !cmd.Flags().Changed("summary") && !cmd.Flags().Changed("due") &&
		!cmd.Flags().Changed("description") && !cmd.Flags().Changed("status") {
		return fmt.Errorf("at least one of --summary, --due, --description, or --status must be specified")
	}

	data := map[string]interface{}{
		"entity_id": entityID,
		"item":      uid,
	}
	if todoUpdateSummary != "" {
		data["rename"] = todoUpdateSummary
	}
	if todoUpdateDescription != "" {
		data["description"] = todoUpdateDescription
	}
	if todoUpdateStatus != "" {
		data["status"] = todoUpdateStatus
	}
	if todoUpdateDue != "" {
		if len(todoUpdateDue) > 10 {
			data["due_datetime"] = todoUpdateDue
		} else {
			data["due_date"] = todoUpdateDue
		}
	}

	return callServiceAction("todo", "update_item", "Item updated.", data)
}

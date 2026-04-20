package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	todoRemoveForce  bool
	todoRemovePlan   bool
	todoRemoveDryRun bool
)

var todoRemoveCmd = &cobra.Command{
	Use:     "remove <entity_id> <item_uid>",
	Short:   "Remove an item from a to-do list",
	Long:    `Remove an item from a Home Assistant to-do list by its uid.`,
	Example: `  hab todo remove todo.shopping_list abc123`,
	Args:    cobra.ExactArgs(2),
	RunE:    runTodoRemove,
}

func init() {
	todoCmd.AddCommand(todoRemoveCmd)
	todoRemoveCmd.Flags().BoolVarP(&todoRemoveForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(todoRemoveCmd, &todoRemovePlan, &todoRemoveDryRun)
	annotateServiceMutationCommand(todoRemoveCmd, "destructive", "todo_item", []string{"args", "flags"})
}

func runTodoRemove(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "todo")
	uid := args[1]
	textMode := getTextMode()

	if planRequested(todoRemovePlan, todoRemoveDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":  "todo_item",
				"entity_id": entityID,
				"item_uid":  uid,
			},
			Steps: []string{
				"Call Home Assistant service todo.remove_item with target entity_id and item UID.",
				"Return deletion confirmation message.",
			},
			Risks: []string{
				"This operation permanently removes the to-do item.",
			},
			RequiresConfirmation: !todoRemoveForce,
			VerificationCommands: []string{
				fmt.Sprintf("hab todo items %s --json", entityID),
			},
		}, textMode, "todo_item")
		return nil
	}

	if err := confirmAction(todoRemoveForce, fmt.Sprintf("Remove to-do item %s from %s?", uid, entityID), "remove to-do item"); err != nil {
		return err
	}

	return callServiceAction("remove", "todo_item", "todo", "remove_item", "Item removed.", map[string]interface{}{
		"entity_id": entityID,
		"item":      uid,
	})
}

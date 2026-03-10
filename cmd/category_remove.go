package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	categoryRemoveEntityID string
	categoryRemoveScope    string
)

var categoryRemoveCmd = &cobra.Command{
	Use:   "remove <entity_id>",
	Short: "Remove a category from an entity",
	Long: `Remove the category assignment from an entity for a given scope.

The scope is inferred from the entity_id prefix when possible.
Use --scope to override.`,
	Example: `  hab category remove automation.evening_lights
  hab category remove input_boolean.guest_mode --scope helpers`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCategoryRemove,
}

func init() {
	categoryCmd.AddCommand(categoryRemoveCmd)
	categoryRemoveCmd.Flags().StringVar(&categoryRemoveEntityID, "entity", "", "Entity ID to remove the category from")
	categoryRemoveCmd.Flags().StringVar(&categoryRemoveScope, "scope", "", "Scope to remove category from: automation, script, scene, helpers")
}

func runCategoryRemove(cmd *cobra.Command, args []string) error {
	entityID, err := resolveArg(categoryRemoveEntityID, args, 0, "entity ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	scope := categoryRemoveScope
	if scope == "" {
		scope = inferCategoryScope(entityID)
	}
	if scope == "" {
		return fmt.Errorf("cannot infer scope from entity_id '%s'. Use --scope to specify: automation, script, scene, helpers", entityID)
	}
	if !validCategoryScopes[scope] {
		return fmt.Errorf("invalid scope '%s'. Valid values: automation, script, scene, helpers", scope)
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	// Set category for this scope to null to remove the assignment
	result, err := ws.EntityRegistryUpdate(entityID, map[string]interface{}{
		"categories": map[string]interface{}{
			scope: nil,
		},
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Category removed from '%s' (scope: %s).", entityID, scope))
	return nil
}

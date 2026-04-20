package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	categoryRemoveEntityID string
	categoryRemoveScope    string
	categoryRemoveForce    bool
	categoryRemovePlan     bool
	categoryRemoveDryRun   bool
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
	categoryRemoveCmd.Flags().StringVar(&categoryRemoveEntityID, "entity-id", "", "Alias for --entity")
	categoryRemoveCmd.Flags().StringVar(&categoryRemoveScope, "scope", "", "Scope to remove category from: automation, script, scene, helpers")
	categoryRemoveCmd.Flags().BoolVarP(&categoryRemoveForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(categoryRemoveCmd, &categoryRemovePlan, &categoryRemoveDryRun)
	mergeSchemaAnnotation(categoryRemoveCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "category_assignment",
		InputSources: []string{"args", "flags"},
	})
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
		noteResolution("category_scope", "inferred")
	} else {
		noteResolution("category_scope", "flag")
	}
	if scope == "" {
		return fmt.Errorf("cannot infer scope from entity_id '%s'. Use --scope to specify: automation, script, scene, helpers", entityID)
	}
	if !validCategoryScopes[scope] {
		return fmt.Errorf("invalid scope '%s'. Valid values: automation, script, scene, helpers", scope)
	}

	if planRequested(categoryRemovePlan, categoryRemoveDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":  "category_assignment",
				"entity_id": entityID,
				"scope":     scope,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Set categories.<scope> to null on the target entity registry entry.",
				"Return updated entity registry entry.",
			},
			Risks: []string{
				"Removing a category changes UI and organizational grouping.",
			},
			RequiresConfirmation: !categoryRemoveForce,
			VerificationCommands: []string{
				fmt.Sprintf("hab category list --scope %s --json", scope),
			},
		}, textMode, "category_assignment")
		return nil
	}

	if err := confirmAction(categoryRemoveForce, fmt.Sprintf("Remove category from '%s' in scope '%s'?", entityID, scope), "remove category assignment"); err != nil {
		return err
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

	output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("Category removed from '%s' (scope: %s).", entityID, scope), output.EnvelopeContext{Operation: "remove", ResourceType: "category_assignment"})
	return nil
}

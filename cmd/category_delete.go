package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	categoryDeleteScope string
	categoryDeleteForce bool
	categoryDeletePlan  bool
	categoryDeleteDry   bool
)

var categoryDeleteCmd = &cobra.Command{
	Use:   "delete <category_id>",
	Short: "Delete a category",
	Long:  `Delete a category from a given scope.`,
	Example: `  hab category delete abc123 --scope automation
  hab category delete abc123 --scope automation --force`,
	Args: cobra.ExactArgs(1),
	RunE: runCategoryDelete,
}

func init() {
	categoryCmd.AddCommand(categoryDeleteCmd)
	categoryDeleteCmd.Flags().StringVar(&categoryDeleteScope, "scope", "", "Scope of the category: automation, script, scene, helpers (required)")
	categoryDeleteCmd.Flags().BoolVarP(&categoryDeleteForce, "force", "f", false, "Skip confirmation")
	categoryDeleteCmd.MarkFlagRequired("scope")
	registerMutationPlanFlags(categoryDeleteCmd, &categoryDeletePlan, &categoryDeleteDry)
	mergeSchemaAnnotation(categoryDeleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "category",
		InputSources: []string{"args", "flags"},
	})
}

func runCategoryDelete(cmd *cobra.Command, args []string) error {
	categoryID := args[0]

	if !validCategoryScopes[categoryDeleteScope] {
		return fmt.Errorf("invalid scope '%s'. Valid values: automation, script, scene, helpers", categoryDeleteScope)
	}
	textMode := getTextMode()

	if planRequested(categoryDeletePlan, categoryDeleteDry) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":    "category",
				"scope":       categoryDeleteScope,
				"category_id": categoryID,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send category registry delete in the selected scope.",
				"Return deletion confirmation.",
			},
			Risks: []string{
				"Deleting a category affects organization for entities in that scope.",
			},
			RequiresConfirmation: !categoryDeleteForce,
			VerificationCommands: []string{
				fmt.Sprintf("hab category list --scope %s --json", categoryDeleteScope),
			},
		}, textMode, "category")
		return nil
	}

	if err := confirmAction(categoryDeleteForce, fmt.Sprintf("Delete category %s from scope %s?", categoryID, categoryDeleteScope), "delete category"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.CategoryRegistryDelete(categoryDeleteScope, categoryID); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Category '%s' deleted.", categoryID), output.EnvelopeContext{Operation: "delete", ResourceType: "category"})
	return nil
}

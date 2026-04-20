package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var helperDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id>",
	Short: "Delete any helper by entity ID",
	Long: `Delete any helper by its entity ID.

This command automatically detects the helper type from the entity ID and deletes it.
Supports: input_boolean, input_number, input_text, input_select, input_datetime,
input_button, counter, timer, schedule, and group helpers.`,
	Example: `  hab helper delete input_boolean.my_toggle
  hab helper delete counter.page_views
  hab helper delete light.living_room_group`,
	GroupID: helperGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runHelperDelete,
}

var (
	helperDeleteForce  bool
	helperDeletePlan   bool
	helperDeleteDryRun bool
)

func init() {
	helperCmd.AddCommand(helperDeleteCmd)
	helperDeleteCmd.Flags().BoolVarP(&helperDeleteForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(helperDeleteCmd, &helperDeletePlan, &helperDeleteDryRun)
	mergeSchemaAnnotation(helperDeleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "helper",
		InputSources: []string{"args", "flags"},
	})
}

func runHelperDelete(cmd *cobra.Command, args []string) error {
	entityID := args[0]

	textMode := getTextMode()

	// Extract domain from entity_id
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid entity_id format: %s (expected domain.object_id)", entityID)
	}
	domain := parts[0]

	// Map domains to helper types
	helperType := ""

	switch domain {
	case "input_boolean", "input_number", "input_text", "input_select", "input_datetime", "input_button":
		helperType = domain
	case "counter":
		helperType = "counter"
	case "timer":
		helperType = "timer"
	case "schedule":
		helperType = "schedule"
	case "light", "switch", "binary_sensor", "cover", "fan", "lock", "media_player", "sensor", "event":
		// These could be group helpers (config entry based)
		helperType = "group"
	default:
		return fmt.Errorf("unsupported helper domain: %s", domain)
	}

	if planRequested(helperDeletePlan, helperDeleteDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":  "helper",
				"entity_id": entityID,
				"type":      helperType,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Resolve helper deletion path (storage helper or config-entry helper).",
				"Submit helper delete command and return success metadata.",
			},
			Risks: []string{
				"This operation permanently removes the helper entity/configuration.",
			},
			RequiresConfirmation: !helperDeleteForce,
			VerificationCommands: []string{
				"hab helper list --json",
			},
		}, textMode, "helper")
		return nil
	}

	if err := confirmAction(helperDeleteForce, fmt.Sprintf("Delete helper %s?", entityID), "delete helper"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	// deleteHelperByEntityOrEntryID handles both storage-based (WS) and
	// config-entry-based helpers.
	err = deleteHelperByEntityOrEntryID(ws, entityID, helperType)
	if err != nil {
		return fmt.Errorf("failed to delete helper: %w", err)
	}

	result := map[string]interface{}{
		"entity_id": entityID,
		"type":      helperType,
		"deleted":   true,
	}

	output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("Helper '%s' deleted successfully.", entityID), output.EnvelopeContext{Operation: "delete", ResourceType: "helper"})
	return nil
}

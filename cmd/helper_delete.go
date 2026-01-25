package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id>",
	Short: "Delete any helper by entity ID",
	Long: `Delete any helper by its entity ID.

This command automatically detects the helper type from the entity ID and deletes it.
Supports: input_boolean, input_number, input_text, input_select, input_datetime,
input_button, counter, timer, schedule, and group helpers.

Examples:
  hab helper delete input_boolean.my_toggle
  hab helper delete counter.page_views
  hab helper delete light.living_room_group`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperDelete,
}

func init() {
	helperCmd.AddCommand(helperDeleteCmd)
}

func runHelperDelete(cmd *cobra.Command, args []string) error {
	entityID := args[0]

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Extract domain from entity_id
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid entity_id format: %s (expected domain.object_id)", entityID)
	}
	domain := parts[0]

	// Map domains to helper types
	helperType := ""
	isConfigEntry := false

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
		isConfigEntry = true
	default:
		return fmt.Errorf("unsupported helper domain: %s", domain)
	}

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	if isConfigEntry {
		// For config entry based helpers (groups), we need to resolve to config entry
		err = ws.DeleteHelperByEntityOrEntryID(entityID, helperType)
	} else {
		// For storage-based helpers, use the WebSocket delete command
		err = ws.DeleteHelperByEntityOrEntryID(entityID, helperType)
	}

	if err != nil {
		return fmt.Errorf("failed to delete helper: %w", err)
	}

	result := map[string]interface{}{
		"entity_id": entityID,
		"type":      helperType,
		"deleted":   true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Helper '%s' deleted successfully.", entityID))
	return nil
}

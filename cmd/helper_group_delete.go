package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperGroupDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id_or_entry_id>",
	Short: "Delete a group",
	Long: `Delete a group helper by its entity ID or config entry ID.

Examples:
  hab helper-group delete light.living_room_lights
  hab helper-group delete abc123def456`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperGroupDelete,
}

func init() {
	helperGroupParentCmd.AddCommand(helperGroupDeleteCmd)
}

func runHelperGroupDelete(cmd *cobra.Command, args []string) error {
	id := args[0]

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	// Use REST API for config entry operations
	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	// If it looks like an entity_id, we need to resolve it to config_entry_id
	entryID := id
	if strings.Contains(id, ".") {
		// It's an entity_id, need to resolve to config entry
		ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
		if err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		resolved, err := ws.ResolveEntityToConfigEntry(id)
		if err != nil {
			return fmt.Errorf("failed to resolve entity_id: %w", err)
		}
		if resolved == "" {
			return fmt.Errorf("entity %s does not have a config entry", id)
		}
		entryID = resolved
	}

	// Delete the config entry via REST API
	err = rest.ConfigEntryDelete(entryID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Group '%s' deleted successfully.", id))
	return nil
}

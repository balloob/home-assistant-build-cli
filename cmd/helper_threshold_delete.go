package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperThresholdDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id_or_entry_id>",
	Short: "Delete a threshold sensor",
	Long: `Delete a threshold binary sensor helper by its entity ID or config entry ID.

Examples:
  hab helper-threshold delete binary_sensor.high_temperature
  hab helper-threshold delete abc123def456`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperThresholdDelete,
}

func init() {
	helperThresholdParentCmd.AddCommand(helperThresholdDeleteCmd)
}

func runHelperThresholdDelete(cmd *cobra.Command, args []string) error {
	id := args[0]

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	rest := client.NewRestClient(creds.URL, creds.AccessToken)

	entryID := id
	if strings.Contains(id, ".") {
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

	err = rest.ConfigEntryDelete(entryID)
	if err != nil {
		return fmt.Errorf("failed to delete threshold sensor: %w", err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Threshold sensor '%s' deleted successfully.", id))
	return nil
}

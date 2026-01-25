package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var searchRelatedCmd = &cobra.Command{
	Use:   "related <item_type> <item_id>",
	Short: "Find related items for any item type",
	Long: `Find all items related to a given item.

Supported item types:
  - entity: Find items related to an entity (e.g., search related entity light.kitchen)
  - device: Find items related to a device
  - area: Find items related to an area
  - floor: Find items related to a floor
  - label: Find items related to a label
  - automation: Find items related to an automation
  - scene: Find items related to a scene
  - script: Find items related to a script
  - config_entry: Find items related to a config entry
  - group: Find items related to a group

Returns related items grouped by type: areas, automations, config_entries, devices, entities, groups, scenes, scripts, etc.`,
	Args: cobra.ExactArgs(2),
	RunE: runSearchRelated,
}

func init() {
	searchCmd.AddCommand(searchRelatedCmd)
}

func runSearchRelated(cmd *cobra.Command, args []string) error {
	itemType := args[0]
	itemID := args[1]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Validate item type
	validTypes := map[string]bool{
		"entity":               true,
		"device":               true,
		"area":                 true,
		"floor":                true,
		"label":                true,
		"automation":           true,
		"automation_blueprint": true,
		"scene":                true,
		"script":               true,
		"script_blueprint":     true,
		"config_entry":         true,
		"group":                true,
	}

	if !validTypes[itemType] {
		return fmt.Errorf("invalid item type '%s'. Valid types: entity, device, area, floor, label, automation, scene, script, config_entry, group", itemType)
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

	related, err := ws.SearchRelated(itemType, itemID)
	if err != nil {
		return err
	}

	// Build result with item info and related items
	result := map[string]interface{}{
		"item_type": itemType,
		"item_id":   itemID,
		"related":   related,
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

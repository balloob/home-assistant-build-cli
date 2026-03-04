package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	searchRelatedType string
	searchRelatedID   string
)

var searchRelatedCmd = &cobra.Command{
	Use:   "related [item_type] [item_id]",
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
	Args: cobra.MaximumNArgs(2),
	RunE: runSearchRelated,
}

func init() {
	searchCmd.AddCommand(searchRelatedCmd)
	searchRelatedCmd.Flags().StringVar(&searchRelatedType, "type", "", "Item type (entity, device, area, floor, label, automation, scene, script, config_entry, group)")
	searchRelatedCmd.Flags().StringVar(&searchRelatedID, "id", "", "Item ID to search for related items")
}

func runSearchRelated(cmd *cobra.Command, args []string) error {
	itemType, err := resolveArg(searchRelatedType, args, 0, "item type")
	if err != nil {
		return err
	}
	itemID, err := resolveArg(searchRelatedID, args, 1, "item ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

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

	ws, err := getWSClient()
	if err != nil {
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

	output.PrintOutput(result, textMode, "")
	return nil
}

package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	searchRelatedType string
	searchRelatedID   string
)

// searchRelatedItemTypes matches ItemType in the Home Assistant search integration.
var searchRelatedItemTypes = []string{
	"area",
	"automation",
	"automation_blueprint",
	"config_entry",
	"device",
	"entity",
	"floor",
	"group",
	"integration",
	"label",
	"person",
	"scene",
	"script",
	"script_blueprint",
}

var searchRelatedCmd = &cobra.Command{
	Use:   "related [item_type] [item_id]",
	Short: "Find related items for any item type",
	Long: `Find all items related to a given item.

Supported item types:
  - area: Find items related to an area (area ID)
  - automation: Find items related to an automation (entity ID)
  - automation_blueprint: Find automations that use a blueprint (blueprint path)
  - config_entry: Find items related to a config entry (entry ID)
  - device: Find items related to a device (device ID)
  - entity: Find items related to an entity (entity ID)
  - floor: Find items related to a floor (floor ID)
  - group: Find items related to a group (entity ID)
  - integration: Find items related to an integration (integration domain)
  - label: Find items related to a label (label ID)
  - person: Find items related to a person (entity ID)
  - scene: Find items related to a scene (entity ID)
  - script: Find items related to a script (entity ID)
  - script_blueprint: Find scripts that use a blueprint (blueprint path)

Returns related item IDs grouped by the same item type names, for example area, automation, config_entry, device, entity, group, integration, scene, and script.`,
	Example: `  hab search related entity light.kitchen
  hab search related area living_room
  hab search related --type device --id abc123def456`,
	Args: cobra.MaximumNArgs(2),
	RunE: runSearchRelated,
}

func init() {
	searchCmd.AddCommand(searchRelatedCmd)
	searchRelatedCmd.Flags().StringVar(&searchRelatedType, "type", "", "Item type ("+strings.Join(searchRelatedItemTypes, ", ")+")")
	searchRelatedCmd.Flags().StringVar(&searchRelatedID, "id", "", "Item ID to search for related items")
	searchRelatedCmd.Flags().StringVar(&searchRelatedType, "item-type", "", "Alias for --type")
	searchRelatedCmd.Flags().StringVar(&searchRelatedID, "item-id", "", "Alias for --id")
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

	if !slices.Contains(searchRelatedItemTypes, itemType) {
		return fmt.Errorf("invalid item type '%s'. Valid types: %s", itemType, strings.Join(searchRelatedItemTypes, ", "))
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

	output.PrintOutputWithContext(result, textMode, "", output.EnvelopeContext{Operation: "search_related", ResourceType: itemType})
	return nil
}

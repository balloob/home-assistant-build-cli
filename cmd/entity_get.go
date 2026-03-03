package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	entityGetRelated bool
	entityGetDevice  bool
	entityGetID      string
)

var entityGetCmd = &cobra.Command{
	Use:   "get [entity_id]",
	Short: "Get entity state, attributes, and registry data",
	Long: `Get the current state, attributes, and registry data of an entity.

Use --related to show related automations, scripts, scenes, and devices.
Use --device to include the parent device information.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEntityGet,
}

func init() {
	entityCmd.AddCommand(entityGetCmd)
	entityGetCmd.Flags().StringVar(&entityGetID, "entity", "", "Entity ID to get")
	entityGetCmd.Flags().BoolVarP(&entityGetRelated, "related", "r", false, "Include related items (automations, scripts, scenes, devices)")
	entityGetCmd.Flags().BoolVarP(&entityGetDevice, "device", "D", false, "Include parent device information")
}

func runEntityGet(cmd *cobra.Command, args []string) error {
	entityID := entityGetID
	if entityID == "" && len(args) > 0 {
		entityID = args[0]
	}
	if entityID == "" {
		return fmt.Errorf("entity ID is required (use --entity flag or positional argument)")
	}
	textMode := getTextMode()

	// Get state from REST API
	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	state, err := restClient.GetState(entityID)
	if err != nil {
		return err
	}

	// Get registry data and optionally related items via WebSocket
	ws, err := getWSClient()
	if err != nil {
		// Fall back to just state if we can't get WebSocket connection
		output.PrintOutput(state, textMode, "")
		return nil
	}
	defer ws.Close()

	// Get entity registry data
	registry, err := ws.EntityRegistryGet(entityID)
	if err != nil {
		// Entity might not be in registry, just return state
		output.PrintOutput(state, textMode, "")
		return nil
	}

	// Build result combining state and registry data
	result := make(map[string]interface{})

	// Copy state data
	for k, v := range state {
		result[k] = v
	}

	// Add registry data under "registry" key
	result["registry"] = registry

	// Get parent device if requested
	if entityGetDevice {
		if deviceID, ok := registry["device_id"].(string); ok && deviceID != "" {
			devices, err := ws.DeviceRegistryList()
			if err == nil {
				for _, d := range devices {
					if device, ok := d.(map[string]interface{}); ok {
						if device["id"] == deviceID {
							result["device"] = device
							break
						}
					}
				}
			}
		}
	}

	// Get related items if requested
	if entityGetRelated {
		related, err := ws.SearchRelated("entity", entityID)
		if err == nil && len(related) > 0 {
			result["related"] = related
		}
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

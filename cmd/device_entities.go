package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var deviceEntitiesID string

var deviceEntitiesCmd = &cobra.Command{
	Use:   "entities [device_id]",
	Short: "List entities for a device",
	Long:  `List all entities that belong to a device.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDeviceEntities,
}

func init() {
	deviceCmd.AddCommand(deviceEntitiesCmd)
	deviceEntitiesCmd.Flags().StringVar(&deviceEntitiesID, "device", "", "Device ID to list entities for")
}

func runDeviceEntities(cmd *cobra.Command, args []string) error {
	deviceID, err := resolveArg(deviceEntitiesID, args, 0, "device ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	entities, err := ws.EntityRegistryList()
	if err != nil {
		return err
	}

	var result []map[string]interface{}
	for _, e := range entities {
		entity, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		if entity["device_id"] == deviceID {
			result = append(result, map[string]interface{}{
				"entity_id": entity["entity_id"],
				"name":      entity["name"],
				"disabled":  entity["disabled_by"] != nil,
			})
		}
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

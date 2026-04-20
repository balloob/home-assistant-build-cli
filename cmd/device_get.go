package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	deviceGetRelated bool
	deviceGetID      string
)

var deviceGetCmd = &cobra.Command{
	Use:   "get [device_id]",
	Short: "Get device details",
	Long:  `Get detailed information about a device. Use --related to also show related automations, scripts, scenes, and entities.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDeviceGet,
}

func init() {
	deviceCmd.AddCommand(deviceGetCmd)
	deviceGetCmd.Flags().StringVar(&deviceGetID, "device", "", "Device ID to get")
	deviceGetCmd.Flags().StringVar(&deviceGetID, "device-id", "", "Alias for --device")
	deviceGetCmd.Flags().BoolVarP(&deviceGetRelated, "related", "r", false, "Include related items (automations, scripts, scenes, entities)")
}

func runDeviceGet(cmd *cobra.Command, args []string) error {
	deviceID, err := resolveArg(deviceGetID, args, 0, "device ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	devices, err := ws.DeviceRegistryList()
	if err != nil {
		return err
	}

	for _, d := range devices {
		device, ok := d.(map[string]interface{})
		if !ok {
			continue
		}
		if device["id"] == deviceID {
			result := device

			// Get related items if requested
			if deviceGetRelated {
				related, err := ws.SearchRelated("device", deviceID)
				if err == nil && len(related) > 0 {
					// Create a new map to avoid modifying the original
					resultMap := make(map[string]interface{})
					for k, v := range device {
						resultMap[k] = v
					}
					resultMap["related"] = related
					result = resultMap
				} else if err != nil {
					noteFallback("device related search failed; returned device registry data without related items")
					noteMissingSection("related")
				}
			}

			output.PrintOutputWithContext(result, textMode, "", output.EnvelopeContext{Operation: "get", ResourceType: "device"})
			return nil
		}
	}

	return fmt.Errorf("device '%s' not found", deviceID)
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	deviceListID    string
	deviceListArea  string
	deviceListFloor string
)

var deviceListFlags *ListFlags

var deviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devices",
	Long:  `List all devices in Home Assistant. Use --area to filter by area, or --floor to filter by floor.`,
	RunE:  runDeviceList,
}

func init() {
	deviceCmd.AddCommand(deviceListCmd)
	deviceListCmd.Flags().StringVar(&deviceListID, "device-id", "", "Filter by device ID")
	deviceListCmd.Flags().StringVarP(&deviceListArea, "area", "a", "", "Filter by area ID")
	deviceListCmd.Flags().StringVarP(&deviceListFloor, "floor", "f", "", "Filter by floor ID (includes all areas on that floor)")
	deviceListFlags = RegisterListFlags(deviceListCmd, "id")
}

func runDeviceList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	// Build area-to-floor map if floor filter is used
	var areaFloorMap map[string]string
	if deviceListFloor != "" {
		areas, err := ws.AreaRegistryList()
		if err == nil {
			areaFloorMap = buildAreaFloorMap(areas)
		}
	}

	devices, err := ws.DeviceRegistryList()
	if err != nil {
		return err
	}

	var result []map[string]interface{}
	for _, d := range devices {
		device, ok := d.(map[string]interface{})
		if !ok {
			continue
		}

		deviceID, _ := device["id"].(string)
		areaID, _ := device["area_id"].(string)

		// Apply device ID filter
		if deviceListID != "" && deviceID != deviceListID {
			continue
		}

		// Apply area filter
		if deviceListArea != "" {
			if areaID != deviceListArea {
				continue
			}
		}

		// Apply floor filter (check if device's area is on the specified floor)
		if deviceListFloor != "" {
			if areaID == "" {
				continue
			}
			floorID := areaFloorMap[areaID]
			if floorID != deviceListFloor {
				continue
			}
		}

		result = append(result, map[string]interface{}{
			"id":           device["id"],
			"name":         device["name"],
			"manufacturer": device["manufacturer"],
			"model":        device["model"],
			"area_id":      device["area_id"],
		})
	}

	if deviceListFlags.RenderCount(len(result), textMode) {
		return nil
	}
	result = deviceListFlags.ApplyLimitMap(result)
	if deviceListFlags.RenderBriefMap(result, textMode, "id", "name") {
		return nil
	}

	// Full output
	if textMode {
		if len(result) == 0 {
			fmt.Println("No devices.")
			return nil
		}
		for _, item := range result {
			name, _ := item["name"].(string)
			id, _ := item["id"].(string)
			manufacturer, _ := item["manufacturer"].(string)
			model, _ := item["model"].(string)
			areaID, _ := item["area_id"].(string)

			fmt.Printf("%s (%s):\n", name, id)
			if manufacturer != "" || model != "" {
				if manufacturer != "" && model != "" {
					fmt.Printf("  %s %s\n", manufacturer, model)
				} else if manufacturer != "" {
					fmt.Printf("  %s\n", manufacturer)
				} else {
					fmt.Printf("  %s\n", model)
				}
			}
			if areaID != "" {
				fmt.Printf("  area: %s\n", areaID)
			}
		}
	} else {
		output.PrintOutput(result, false, "")
	}
	return nil
}

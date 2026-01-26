package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	deviceListID    string
	deviceListArea  string
	deviceListFloor string
	deviceListCount bool
	deviceListBrief bool
	deviceListLimit int
)

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
	deviceListCmd.Flags().BoolVarP(&deviceListCount, "count", "c", false, "Return only the count of items")
	deviceListCmd.Flags().BoolVarP(&deviceListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	deviceListCmd.Flags().IntVarP(&deviceListLimit, "limit", "n", 0, "Limit results to N items")
}

func runDeviceList(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

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

	// Build area-to-floor map if floor filter is used
	var areaFloorMap map[string]string
	if deviceListFloor != "" {
		areaFloorMap = make(map[string]string)
		areas, err := ws.AreaRegistryList()
		if err == nil {
			for _, a := range areas {
				if area, ok := a.(map[string]interface{}); ok {
					areaID, _ := area["area_id"].(string)
					floorID, _ := area["floor_id"].(string)
					if areaID != "" {
						areaFloorMap[areaID] = floorID
					}
				}
			}
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

	// Handle count mode
	if deviceListCount {
		if textMode {
			fmt.Printf("Count: %d\n", len(result))
		} else {
			client.PrintOutput(map[string]interface{}{"count": len(result)}, false, "")
		}
		return nil
	}

	// Apply limit
	if deviceListLimit > 0 && len(result) > deviceListLimit {
		result = result[:deviceListLimit]
	}

	// Handle brief mode
	if deviceListBrief {
		if textMode {
			for _, item := range result {
				name, _ := item["name"].(string)
				id, _ := item["id"].(string)
				fmt.Printf("%s (%s)\n", name, id)
			}
		} else {
			var brief []map[string]interface{}
			for _, item := range result {
				brief = append(brief, map[string]interface{}{
					"id":   item["id"],
					"name": item["name"],
				})
			}
			client.PrintOutput(brief, false, "")
		}
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
		client.PrintOutput(result, false, "")
	}
	return nil
}

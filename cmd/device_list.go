package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	deviceListArea  string
	deviceListFloor string
)

var deviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devices",
	Long:  `List all devices in Home Assistant. Use --area to filter by area, or --floor to filter by floor.`,
	RunE:  runDeviceList,
}

func init() {
	deviceCmd.AddCommand(deviceListCmd)
	deviceListCmd.Flags().StringVarP(&deviceListArea, "area", "a", "", "Filter by area ID")
	deviceListCmd.Flags().StringVarP(&deviceListFloor, "floor", "f", "", "Filter by floor ID (includes all areas on that floor)")
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

		areaID, _ := device["area_id"].(string)

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

	client.PrintOutput(result, textMode, "")
	return nil
}

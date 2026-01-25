package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deviceListArea string

var deviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devices",
	Long:  `List all devices in Home Assistant. Use --area to filter by area.`,
	RunE:  runDeviceList,
}

func init() {
	deviceCmd.AddCommand(deviceListCmd)
	deviceListCmd.Flags().StringVarP(&deviceListArea, "area", "a", "", "Filter by area ID")
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

		// Apply area filter
		if deviceListArea != "" {
			areaID, _ := device["area_id"].(string)
			if areaID != deviceListArea {
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

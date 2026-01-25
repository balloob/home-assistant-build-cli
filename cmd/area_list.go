package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var areaListFloor string

var areaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all areas",
	Long:  `List all areas in Home Assistant. Use --floor to filter by floor.`,
	RunE:  runAreaList,
}

func init() {
	areaCmd.AddCommand(areaListCmd)
	areaListCmd.Flags().StringVarP(&areaListFloor, "floor", "f", "", "Filter by floor ID")
}

func runAreaList(cmd *cobra.Command, args []string) error {
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

	areas, err := ws.AreaRegistryList()
	if err != nil {
		return err
	}

	var result []map[string]interface{}
	for _, a := range areas {
		area, ok := a.(map[string]interface{})
		if !ok {
			continue
		}

		// Apply floor filter
		if areaListFloor != "" {
			floorID, _ := area["floor_id"].(string)
			if floorID != areaListFloor {
				continue
			}
		}

		result = append(result, map[string]interface{}{
			"area_id":  area["area_id"],
			"name":     area["name"],
			"floor_id": area["floor_id"],
			"icon":     area["icon"],
			"labels":   area["labels"],
		})
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

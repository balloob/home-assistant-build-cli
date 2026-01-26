package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	areaListID    string
	areaListFloor string
	areaListCount bool
	areaListBrief bool
	areaListLimit int
)

var areaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all areas",
	Long:  `List all areas in Home Assistant. Use --floor to filter by floor.`,
	RunE:  runAreaList,
}

func init() {
	areaCmd.AddCommand(areaListCmd)
	areaListCmd.Flags().StringVar(&areaListID, "area-id", "", "Filter by area ID")
	areaListCmd.Flags().StringVarP(&areaListFloor, "floor", "f", "", "Filter by floor ID")
	areaListCmd.Flags().BoolVarP(&areaListCount, "count", "c", false, "Return only the count of items")
	areaListCmd.Flags().BoolVarP(&areaListBrief, "brief", "b", false, "Return minimal fields (area_id and name only)")
	areaListCmd.Flags().IntVarP(&areaListLimit, "limit", "n", 0, "Limit results to N items")
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

		areaID, _ := area["area_id"].(string)

		// Apply area ID filter
		if areaListID != "" && areaID != areaListID {
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

	// Handle count mode
	if areaListCount {
		if textMode {
			fmt.Printf("Count: %d\n", len(result))
		} else {
			client.PrintOutput(map[string]interface{}{"count": len(result)}, false, "")
		}
		return nil
	}

	// Apply limit
	if areaListLimit > 0 && len(result) > areaListLimit {
		result = result[:areaListLimit]
	}

	// Handle brief mode
	if areaListBrief {
		if textMode {
			for _, item := range result {
				name, _ := item["name"].(string)
				areaID, _ := item["area_id"].(string)
				fmt.Printf("%s (%s)\n", name, areaID)
			}
		} else {
			var brief []map[string]interface{}
			for _, item := range result {
				brief = append(brief, map[string]interface{}{
					"area_id": item["area_id"],
					"name":    item["name"],
				})
			}
			client.PrintOutput(brief, false, "")
		}
		return nil
	}

	// Full output
	if textMode {
		if len(result) == 0 {
			fmt.Println("No areas.")
			return nil
		}
		for _, item := range result {
			name, _ := item["name"].(string)
			areaID, _ := item["area_id"].(string)
			floorID, _ := item["floor_id"].(string)

			fmt.Printf("%s (%s):\n", name, areaID)
			if floorID != "" {
				fmt.Printf("  floor: %s\n", floorID)
			}
		}
	} else {
		client.PrintOutput(result, false, "")
	}
	return nil
}

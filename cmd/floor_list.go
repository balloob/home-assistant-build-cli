package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var floorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all floors",
	Long:  `List all floors in Home Assistant.`,
	RunE:  runFloorList,
}

var (
	floorListCount bool
	floorListBrief bool
	floorListLimit int
)

func init() {
	floorCmd.AddCommand(floorListCmd)
	floorListCmd.Flags().BoolVarP(&floorListCount, "count", "c", false, "Return only the count of items")
	floorListCmd.Flags().BoolVarP(&floorListBrief, "brief", "b", false, "Return minimal fields (floor_id and name only)")
	floorListCmd.Flags().IntVarP(&floorListLimit, "limit", "n", 0, "Limit results to N items")
}

func runFloorList(cmd *cobra.Command, args []string) error {
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

	floors, err := ws.FloorRegistryList()
	if err != nil {
		return err
	}

	// Handle count mode
	if floorListCount {
		if textMode {
			fmt.Printf("Count: %d\n", len(floors))
		} else {
			client.PrintOutput(map[string]interface{}{"count": len(floors)}, false, "")
		}
		return nil
	}

	// Apply limit
	if floorListLimit > 0 && len(floors) > floorListLimit {
		floors = floors[:floorListLimit]
	}

	// Handle brief mode
	if floorListBrief {
		if textMode {
			for _, f := range floors {
				if floor, ok := f.(map[string]interface{}); ok {
					name, _ := floor["name"].(string)
					floorID, _ := floor["floor_id"].(string)
					fmt.Printf("%s (%s)\n", name, floorID)
				}
			}
		} else {
			var brief []map[string]interface{}
			for _, f := range floors {
				if floor, ok := f.(map[string]interface{}); ok {
					brief = append(brief, map[string]interface{}{
						"floor_id": floor["floor_id"],
						"name":     floor["name"],
					})
				}
			}
			client.PrintOutput(brief, false, "")
		}
		return nil
	}

	// Full output
	if textMode {
		if len(floors) == 0 {
			fmt.Println("No floors.")
			return nil
		}
		for _, f := range floors {
			if floor, ok := f.(map[string]interface{}); ok {
				name, _ := floor["name"].(string)
				floorID, _ := floor["floor_id"].(string)
				level, hasLevel := floor["level"].(float64)

				if hasLevel {
					fmt.Printf("%s (%s): level %.0f\n", name, floorID, level)
				} else {
					fmt.Printf("%s (%s)\n", name, floorID)
				}
			}
		}
	} else {
		client.PrintOutput(floors, false, "")
	}
	return nil
}

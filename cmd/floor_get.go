package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var floorGetRelated bool

var floorGetCmd = &cobra.Command{
	Use:   "get <floor_id>",
	Short: "Get floor details",
	Long:  `Get detailed information about a floor. Use --related to also show related areas, devices, entities, automations, and scripts.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runFloorGet,
}

func init() {
	floorCmd.AddCommand(floorGetCmd)
	floorGetCmd.Flags().BoolVarP(&floorGetRelated, "related", "r", false, "Include related items (areas, devices, entities, automations, scripts)")
}

func runFloorGet(cmd *cobra.Command, args []string) error {
	floorID := args[0]
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

	for _, f := range floors {
		floor, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		if floor["floor_id"] == floorID {
			result := floor

			// Get related items if requested
			if floorGetRelated {
				related, err := ws.SearchRelated("floor", floorID)
				if err == nil && len(related) > 0 {
					// Create a new map to avoid modifying the original
					resultMap := make(map[string]interface{})
					for k, v := range floor {
						resultMap[k] = v
					}
					resultMap["related"] = related
					result = resultMap
				}
			}

			client.PrintOutput(result, textMode, "")
			return nil
		}
	}

	return fmt.Errorf("floor '%s' not found", floorID)
}

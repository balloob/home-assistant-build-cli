package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	areaGetRelated bool
	areaGetID      string
)

var areaGetCmd = &cobra.Command{
	Use:   "get [area_id]",
	Short: "Get area details",
	Long:  `Get detailed information about an area. Use --related to also show related devices, entities, automations, and scripts.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runAreaGet,
}

func init() {
	areaCmd.AddCommand(areaGetCmd)
	areaGetCmd.Flags().StringVar(&areaGetID, "area", "", "Area ID to get")
	areaGetCmd.Flags().BoolVarP(&areaGetRelated, "related", "r", false, "Include related items (devices, entities, automations, scripts)")
}

func runAreaGet(cmd *cobra.Command, args []string) error {
	areaID := areaGetID
	if areaID == "" && len(args) > 0 {
		areaID = args[0]
	}
	if areaID == "" {
		return fmt.Errorf("area ID is required (use --area flag or positional argument)")
	}
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

	for _, a := range areas {
		area, ok := a.(map[string]interface{})
		if !ok {
			continue
		}
		if area["area_id"] == areaID {
			result := area

			// Get related items if requested
			if areaGetRelated {
				related, err := ws.SearchRelated("area", areaID)
				if err == nil && len(related) > 0 {
					// Create a new map to avoid modifying the original
					resultMap := make(map[string]interface{})
					for k, v := range area {
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

	return fmt.Errorf("area '%s' not found", areaID)
}

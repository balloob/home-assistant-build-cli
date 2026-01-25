package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var zoneListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones",
	Long:  `List all zones in Home Assistant.`,
	RunE:  runZoneList,
}

var (
	zoneListCount bool
	zoneListBrief bool
	zoneListLimit int
)

func init() {
	zoneCmd.AddCommand(zoneListCmd)
	zoneListCmd.Flags().BoolVarP(&zoneListCount, "count", "c", false, "Return only the count of items")
	zoneListCmd.Flags().BoolVarP(&zoneListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	zoneListCmd.Flags().IntVarP(&zoneListLimit, "limit", "n", 0, "Limit results to N items")
}

func runZoneList(cmd *cobra.Command, args []string) error {
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

	zones, err := ws.ZoneList()
	if err != nil {
		return err
	}

	// Handle count mode
	if zoneListCount {
		client.PrintOutput(map[string]interface{}{"count": len(zones)}, textMode, "")
		return nil
	}

	// Apply limit
	if zoneListLimit > 0 && len(zones) > zoneListLimit {
		zones = zones[:zoneListLimit]
	}

	// Handle brief mode
	if zoneListBrief {
		var brief []map[string]interface{}
		for _, z := range zones {
			if zone, ok := z.(map[string]interface{}); ok {
				brief = append(brief, map[string]interface{}{
					"id":   zone["id"],
					"name": zone["name"],
				})
			}
		}
		client.PrintOutput(brief, textMode, "")
		return nil
	}

	client.PrintOutput(zones, textMode, "")
	return nil
}

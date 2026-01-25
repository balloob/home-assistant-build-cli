package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperScheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all schedule helpers",
	Long:  `List all schedule helpers.`,
	RunE:  runHelperScheduleList,
}

var (
	helperScheduleListCount bool
	helperScheduleListBrief bool
	helperScheduleListLimit int
)

func init() {
	helperScheduleParentCmd.AddCommand(helperScheduleListCmd)
	helperScheduleListCmd.Flags().BoolVarP(&helperScheduleListCount, "count", "c", false, "Return only the count of items")
	helperScheduleListCmd.Flags().BoolVarP(&helperScheduleListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	helperScheduleListCmd.Flags().IntVarP(&helperScheduleListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperScheduleList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("schedule")
	if err != nil {
		return err
	}

	// Handle count mode
	if helperScheduleListCount {
		client.PrintOutput(map[string]interface{}{"count": len(helpers)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperScheduleListLimit > 0 && len(helpers) > helperScheduleListLimit {
		helpers = helpers[:helperScheduleListLimit]
	}

	// Handle brief mode
	if helperScheduleListBrief {
		var brief []map[string]interface{}
		for _, h := range helpers {
			if helper, ok := h.(map[string]interface{}); ok {
				brief = append(brief, map[string]interface{}{
					"id":   helper["id"],
					"name": helper["name"],
				})
			}
		}
		client.PrintOutput(brief, textMode, "")
		return nil
	}

	client.PrintOutput(helpers, textMode, "")
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperTimerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all timer helpers",
	Long:  `List all timer helpers.`,
	RunE:  runHelperTimerList,
}

var (
	helperTimerListCount bool
	helperTimerListBrief bool
	helperTimerListLimit int
)

func init() {
	helperTimerParentCmd.AddCommand(helperTimerListCmd)
	helperTimerListCmd.Flags().BoolVarP(&helperTimerListCount, "count", "c", false, "Return only the count of items")
	helperTimerListCmd.Flags().BoolVarP(&helperTimerListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	helperTimerListCmd.Flags().IntVarP(&helperTimerListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperTimerList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("timer")
	if err != nil {
		return err
	}

	// Handle count mode
	if helperTimerListCount {
		client.PrintOutput(map[string]interface{}{"count": len(helpers)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperTimerListLimit > 0 && len(helpers) > helperTimerListLimit {
		helpers = helpers[:helperTimerListLimit]
	}

	// Handle brief mode
	if helperTimerListBrief {
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

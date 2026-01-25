package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperCounterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all counter helpers",
	Long:  `List all counter helpers.`,
	RunE:  runHelperCounterList,
}

var (
	helperCounterListCount bool
	helperCounterListBrief bool
	helperCounterListLimit int
)

func init() {
	helperCounterParentCmd.AddCommand(helperCounterListCmd)
	helperCounterListCmd.Flags().BoolVarP(&helperCounterListCount, "count", "c", false, "Return only the count of items")
	helperCounterListCmd.Flags().BoolVarP(&helperCounterListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	helperCounterListCmd.Flags().IntVarP(&helperCounterListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperCounterList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("counter")
	if err != nil {
		return err
	}

	// Handle count mode
	if helperCounterListCount {
		client.PrintOutput(map[string]interface{}{"count": len(helpers)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperCounterListLimit > 0 && len(helpers) > helperCounterListLimit {
		helpers = helpers[:helperCounterListLimit]
	}

	// Handle brief mode
	if helperCounterListBrief {
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

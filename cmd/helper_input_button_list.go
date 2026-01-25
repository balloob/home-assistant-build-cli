package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputButtonListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all input button helpers",
	Long:  `List all input button helpers.`,
	RunE:  runHelperInputButtonList,
}

var (
	helperInputButtonListCount bool
	helperInputButtonListBrief bool
	helperInputButtonListLimit int
)

func init() {
	helperInputButtonParentCmd.AddCommand(helperInputButtonListCmd)
	helperInputButtonListCmd.Flags().BoolVarP(&helperInputButtonListCount, "count", "c", false, "Return only the count of items")
	helperInputButtonListCmd.Flags().BoolVarP(&helperInputButtonListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	helperInputButtonListCmd.Flags().IntVarP(&helperInputButtonListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperInputButtonList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("input_button")
	if err != nil {
		return err
	}

	// Handle count mode
	if helperInputButtonListCount {
		client.PrintOutput(map[string]interface{}{"count": len(helpers)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperInputButtonListLimit > 0 && len(helpers) > helperInputButtonListLimit {
		helpers = helpers[:helperInputButtonListLimit]
	}

	// Handle brief mode
	if helperInputButtonListBrief {
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

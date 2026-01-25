package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputSelectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all input select helpers",
	Long:  `List all input select (dropdown) helpers.`,
	RunE:  runHelperInputSelectList,
}

var (
	helperInputSelectListCount bool
	helperInputSelectListBrief bool
	helperInputSelectListLimit int
)

func init() {
	helperInputSelectParentCmd.AddCommand(helperInputSelectListCmd)
	helperInputSelectListCmd.Flags().BoolVarP(&helperInputSelectListCount, "count", "c", false, "Return only the count of items")
	helperInputSelectListCmd.Flags().BoolVarP(&helperInputSelectListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	helperInputSelectListCmd.Flags().IntVarP(&helperInputSelectListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperInputSelectList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("input_select")
	if err != nil {
		return err
	}

	// Handle count mode
	if helperInputSelectListCount {
		client.PrintOutput(map[string]interface{}{"count": len(helpers)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperInputSelectListLimit > 0 && len(helpers) > helperInputSelectListLimit {
		helpers = helpers[:helperInputSelectListLimit]
	}

	// Handle brief mode
	if helperInputSelectListBrief {
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

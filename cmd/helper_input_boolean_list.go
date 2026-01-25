package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputBooleanListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all input boolean helpers",
	Long:  `List all input boolean (toggle) helpers.`,
	RunE:  runHelperInputBooleanList,
}

var (
	helperInputBooleanListCount bool
	helperInputBooleanListBrief bool
	helperInputBooleanListLimit int
)

func init() {
	helperInputBooleanParentCmd.AddCommand(helperInputBooleanListCmd)
	helperInputBooleanListCmd.Flags().BoolVarP(&helperInputBooleanListCount, "count", "c", false, "Return only the count of items")
	helperInputBooleanListCmd.Flags().BoolVarP(&helperInputBooleanListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	helperInputBooleanListCmd.Flags().IntVarP(&helperInputBooleanListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperInputBooleanList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("input_boolean")
	if err != nil {
		return err
	}

	// Handle count mode
	if helperInputBooleanListCount {
		client.PrintOutput(map[string]interface{}{"count": len(helpers)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperInputBooleanListLimit > 0 && len(helpers) > helperInputBooleanListLimit {
		helpers = helpers[:helperInputBooleanListLimit]
	}

	// Handle brief mode
	if helperInputBooleanListBrief {
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

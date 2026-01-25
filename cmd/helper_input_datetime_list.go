package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputDatetimeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all input datetime helpers",
	Long:  `List all input datetime helpers.`,
	RunE:  runHelperInputDatetimeList,
}

var (
	helperInputDatetimeListCount bool
	helperInputDatetimeListBrief bool
	helperInputDatetimeListLimit int
)

func init() {
	helperInputDatetimeParentCmd.AddCommand(helperInputDatetimeListCmd)
	helperInputDatetimeListCmd.Flags().BoolVarP(&helperInputDatetimeListCount, "count", "c", false, "Return only the count of items")
	helperInputDatetimeListCmd.Flags().BoolVarP(&helperInputDatetimeListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	helperInputDatetimeListCmd.Flags().IntVarP(&helperInputDatetimeListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperInputDatetimeList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("input_datetime")
	if err != nil {
		return err
	}

	// Handle count mode
	if helperInputDatetimeListCount {
		client.PrintOutput(map[string]interface{}{"count": len(helpers)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperInputDatetimeListLimit > 0 && len(helpers) > helperInputDatetimeListLimit {
		helpers = helpers[:helperInputDatetimeListLimit]
	}

	// Handle brief mode
	if helperInputDatetimeListBrief {
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

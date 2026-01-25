package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputTextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all input text helpers",
	Long:  `List all input text helpers.`,
	RunE:  runHelperInputTextList,
}

var (
	helperInputTextListCount bool
	helperInputTextListBrief bool
	helperInputTextListLimit int
)

func init() {
	helperInputTextParentCmd.AddCommand(helperInputTextListCmd)
	helperInputTextListCmd.Flags().BoolVarP(&helperInputTextListCount, "count", "c", false, "Return only the count of items")
	helperInputTextListCmd.Flags().BoolVarP(&helperInputTextListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	helperInputTextListCmd.Flags().IntVarP(&helperInputTextListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperInputTextList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("input_text")
	if err != nil {
		return err
	}

	// Handle count mode
	if helperInputTextListCount {
		client.PrintOutput(map[string]interface{}{"count": len(helpers)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperInputTextListLimit > 0 && len(helpers) > helperInputTextListLimit {
		helpers = helpers[:helperInputTextListLimit]
	}

	// Handle brief mode
	if helperInputTextListBrief {
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

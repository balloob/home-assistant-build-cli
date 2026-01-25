package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputNumberListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all input number helpers",
	Long:  `List all input number helpers.`,
	RunE:  runHelperInputNumberList,
}

var (
	helperInputNumberListCount bool
	helperInputNumberListBrief bool
	helperInputNumberListLimit int
)

func init() {
	helperInputNumberParentCmd.AddCommand(helperInputNumberListCmd)
	helperInputNumberListCmd.Flags().BoolVarP(&helperInputNumberListCount, "count", "c", false, "Return only the count of items")
	helperInputNumberListCmd.Flags().BoolVarP(&helperInputNumberListBrief, "brief", "b", false, "Return minimal fields (id and name only)")
	helperInputNumberListCmd.Flags().IntVarP(&helperInputNumberListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperInputNumberList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("input_number")
	if err != nil {
		return err
	}

	// Handle count mode
	if helperInputNumberListCount {
		client.PrintOutput(map[string]interface{}{"count": len(helpers)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperInputNumberListLimit > 0 && len(helpers) > helperInputNumberListLimit {
		helpers = helpers[:helperInputNumberListLimit]
	}

	// Handle brief mode
	if helperInputNumberListBrief {
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

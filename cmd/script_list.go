package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scriptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all scripts",
	Long:  `List all scripts in Home Assistant.`,
	RunE:  runScriptList,
}

func init() {
	scriptCmd.AddCommand(scriptListCmd)
	scriptListCmd.Flags().BoolP("count", "c", false, "Return only the count of items")
	scriptListCmd.Flags().BoolP("brief", "b", false, "Return minimal fields (entity_id and alias only)")
	scriptListCmd.Flags().IntP("limit", "n", 0, "Limit results to N items")
}

func runScriptList(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")
	listCount, _ := cmd.Flags().GetBool("count")
	listBrief, _ := cmd.Flags().GetBool("brief")
	listLimit, _ := cmd.Flags().GetInt("limit")

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

	states, err := ws.GetStates()
	if err != nil {
		return err
	}

	var result []map[string]interface{}
	for _, s := range states {
		state, ok := s.(map[string]interface{})
		if !ok {
			continue
		}

		entityID, _ := state["entity_id"].(string)
		if !strings.HasPrefix(entityID, "script.") {
			continue
		}

		attrs, _ := state["attributes"].(map[string]interface{})
		result = append(result, map[string]interface{}{
			"entity_id":      entityID,
			"alias":          attrs["friendly_name"],
			"state":          state["state"],
			"last_triggered": attrs["last_triggered"],
		})
	}

	// Handle count mode
	if listCount {
		client.PrintOutput(map[string]interface{}{"count": len(result)}, textMode, "")
		return nil
	}

	// Apply limit
	if listLimit > 0 && len(result) > listLimit {
		result = result[:listLimit]
	}

	// Handle brief mode
	if listBrief {
		var brief []map[string]interface{}
		for _, item := range result {
			brief = append(brief, map[string]interface{}{
				"entity_id": item["entity_id"],
				"alias":     item["alias"],
			})
		}
		client.PrintOutput(brief, textMode, "")
		return nil
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

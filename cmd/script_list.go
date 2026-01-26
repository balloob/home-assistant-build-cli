package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scriptListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all scripts",
	Long:    `List all scripts in Home Assistant.`,
	GroupID: scriptGroupCommands,
	RunE:    runScriptList,
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
		if textMode {
			fmt.Printf("Count: %d\n", len(result))
		} else {
			client.PrintOutput(map[string]interface{}{"count": len(result)}, false, "")
		}
		return nil
	}

	// Apply limit
	if listLimit > 0 && len(result) > listLimit {
		result = result[:listLimit]
	}

	// Handle brief mode
	if listBrief {
		if textMode {
			for _, item := range result {
				alias, _ := item["alias"].(string)
				entityID, _ := item["entity_id"].(string)
				fmt.Printf("%s (%s)\n", alias, entityID)
			}
		} else {
			var brief []map[string]interface{}
			for _, item := range result {
				brief = append(brief, map[string]interface{}{
					"entity_id": item["entity_id"],
					"alias":     item["alias"],
				})
			}
			client.PrintOutput(brief, false, "")
		}
		return nil
	}

	// Full output
	if textMode {
		if len(result) == 0 {
			fmt.Println("No scripts.")
			return nil
		}
		for _, item := range result {
			alias, _ := item["alias"].(string)
			entityID, _ := item["entity_id"].(string)
			state, _ := item["state"].(string)
			lastTriggered, _ := item["last_triggered"].(string)

			fmt.Printf("%s (%s): %s\n", alias, entityID, state)
			if lastTriggered != "" {
				fmt.Printf("  last_triggered: %s\n", lastTriggered)
			}
		}
	} else {
		client.PrintOutput(result, false, "")
	}
	return nil
}

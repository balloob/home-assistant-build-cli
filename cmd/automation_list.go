package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const maxDescriptionLength = 200

var automationListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all automations",
	Long:    `List all automations in Home Assistant.`,
	GroupID: automationGroupCommands,
	RunE:    runAutomationList,
}

func init() {
	automationCmd.AddCommand(automationListCmd)
	automationListCmd.Flags().Bool("extended", false, "Include extended info (description, blueprint) - requires extra API calls")
	automationListCmd.Flags().String("blueprint", "", "Filter to automations using specific blueprint path (implies --extended)")
	automationListCmd.Flags().BoolP("count", "c", false, "Return only the count of items")
	automationListCmd.Flags().BoolP("brief", "b", false, "Return minimal fields (entity_id and alias only)")
	automationListCmd.Flags().IntP("limit", "n", 0, "Limit results to N items")
}

func runAutomationList(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")
	extended, _ := cmd.Flags().GetBool("extended")
	blueprintFilter, _ := cmd.Flags().GetString("blueprint")
	listCount, _ := cmd.Flags().GetBool("count")
	listBrief, _ := cmd.Flags().GetBool("brief")
	listLimit, _ := cmd.Flags().GetInt("limit")

	// Blueprint filter implies extended mode
	if blueprintFilter != "" {
		extended = true
	}

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

	// Get REST client for extended info
	var restClient *client.RestClient
	if extended {
		restClient, err = manager.GetRestClient()
		if err != nil {
			return err
		}
	}

	var result []map[string]interface{}
	for _, s := range states {
		state, ok := s.(map[string]interface{})
		if !ok {
			continue
		}

		entityID, _ := state["entity_id"].(string)
		if !strings.HasPrefix(entityID, "automation.") {
			continue
		}

		attrs, _ := state["attributes"].(map[string]interface{})
		item := map[string]interface{}{
			"entity_id":      entityID,
			"alias":          attrs["friendly_name"],
			"state":          state["state"],
			"last_triggered": attrs["last_triggered"],
		}

		var blueprintPath string

		// Fetch extended info if requested
		if extended && restClient != nil {
			automationID := strings.TrimPrefix(entityID, "automation.")
			config, err := restClient.Get("config/automation/config/" + automationID)
			if err == nil {
				if configMap, ok := config.(map[string]interface{}); ok {
					// Add description (capped)
					if desc, ok := configMap["description"].(string); ok && desc != "" {
						if len(desc) > maxDescriptionLength {
							desc = desc[:maxDescriptionLength] + "..."
						}
						item["description"] = desc
					}
					// Add blueprint info
					if blueprint, ok := configMap["use_blueprint"].(map[string]interface{}); ok {
						if path, ok := blueprint["path"].(string); ok {
							item["blueprint"] = path
							blueprintPath = path
						}
					}
				}
			}
		}

		// Apply blueprint filter
		if blueprintFilter != "" {
			// Filter by specific blueprint path, or use "*" to match any blueprint
			if blueprintFilter == "*" {
				if blueprintPath == "" {
					continue
				}
			} else if blueprintPath != blueprintFilter {
				continue
			}
		}

		result = append(result, item)
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
			fmt.Println("No automations.")
			return nil
		}
		for _, item := range result {
			alias, _ := item["alias"].(string)
			entityID, _ := item["entity_id"].(string)
			state, _ := item["state"].(string)
			lastTriggered, _ := item["last_triggered"].(string)
			description, _ := item["description"].(string)
			blueprint, _ := item["blueprint"].(string)

			fmt.Printf("%s (%s): %s\n", alias, entityID, state)
			if lastTriggered != "" {
				fmt.Printf("  last_triggered: %s\n", lastTriggered)
			}
			if description != "" {
				fmt.Printf("  description: %s\n", description)
			}
			if blueprint != "" {
				fmt.Printf("  blueprint: %s\n", blueprint)
			}
		}
	} else {
		client.PrintOutput(result, false, "")
	}
	return nil
}

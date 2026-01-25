package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const maxDescriptionLength = 200

var automationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all automations",
	Long:  `List all automations in Home Assistant.`,
	RunE:  runAutomationList,
}

func init() {
	automationCmd.AddCommand(automationListCmd)
	automationListCmd.Flags().Bool("extended", false, "Include extended info (description, blueprint) - requires extra API calls")
	automationListCmd.Flags().String("blueprint", "", "Filter to automations using specific blueprint path (implies --extended)")
}

func runAutomationList(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")
	extended, _ := cmd.Flags().GetBool("extended")
	blueprintFilter, _ := cmd.Flags().GetString("blueprint")

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

	client.PrintOutput(result, textMode, "")
	return nil
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

const maxDescriptionLength = 200

var automationListFlags *ListFlags

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
	automationListFlags = RegisterListFlags(automationListCmd, "entity_id")
}

func runAutomationList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()
	extended, _ := cmd.Flags().GetBool("extended")
	blueprintFilter, _ := cmd.Flags().GetString("blueprint")

	// Blueprint filter implies extended mode
	if blueprintFilter != "" {
		extended = true
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	states, err := ws.GetStates()
	if err != nil {
		return err
	}

	// Get REST client for extended info
	var restClient client.RestAPI
	if extended {
		restClient, err = getRESTClient()
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

	if automationListFlags.RenderCount(len(result), textMode) {
		return nil
	}
	result = automationListFlags.ApplyLimitMap(result)
	if automationListFlags.RenderBriefMap(result, textMode, "entity_id", "alias") {
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
		output.PrintOutput(result, false, "")
	}
	return nil
}

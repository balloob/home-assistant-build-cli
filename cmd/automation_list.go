package cmd

import (
	"strings"
	"sync"

	"github.com/home-assistant/hab/client"
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
	extended, _ := cmd.Flags().GetBool("extended")
	blueprintFilter, _ := cmd.Flags().GetString("blueprint")

	// Blueprint filter implies extended mode
	if blueprintFilter != "" {
		extended = true
	}

	// Build enrichment callback for --extended mode.
	var enrichItems func(items []map[string]interface{}) error
	if extended {
		restClient, err := getRESTClient()
		if err != nil {
			return err
		}
		enrichItems = func(items []map[string]interface{}) error {
			return enrichAutomationItems(items, restClient)
		}
	}

	// Build filter callback for --blueprint mode.
	var filterItem func(item map[string]interface{}) bool
	if blueprintFilter != "" {
		filterItem = func(item map[string]interface{}) bool {
			blueprintPath, _ := item["blueprint"].(string)
			if blueprintFilter == "*" {
				return blueprintPath != ""
			}
			return blueprintPath == blueprintFilter
		}
	}

	// Extra text-mode fields when --extended is active.
	var extraTextFields []string
	if extended {
		extraTextFields = []string{"description", "blueprint"}
	}

	return listDomainEntities(stateListConfig{
		domain:          "automation",
		listFlags:       automationListFlags,
		textMode:        getTextMode(),
		emptyMessage:    "No automations.",
		extraTextFields: extraTextFields,
		enrichItems:     enrichItems,
		filterItem:      filterItem,
	})
}

// enrichAutomationItems fetches per-automation configs concurrently and merges
// description and blueprint information into the items.
func enrichAutomationItems(items []map[string]interface{}, restClient client.RestAPI) error {
	const maxConcurrent = 10
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	type extendedInfo struct {
		description string
		blueprint   string
	}
	extInfos := make([]extendedInfo, len(items))

	for i, item := range items {
		entityID, _ := item["entity_id"].(string)
		automationID := strings.TrimPrefix(entityID, "automation.")

		wg.Add(1)
		go func(idx int, autoID string) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			config, err := restClient.Get("config/automation/config/" + autoID)
			if err != nil {
				return
			}
			configMap, ok := config.(map[string]interface{})
			if !ok {
				return
			}

			var info extendedInfo
			if desc, ok := configMap["description"].(string); ok && desc != "" {
				if len(desc) > maxDescriptionLength {
					desc = desc[:maxDescriptionLength] + "..."
				}
				info.description = desc
			}
			if blueprint, ok := configMap["use_blueprint"].(map[string]interface{}); ok {
				if path, ok := blueprint["path"].(string); ok {
					info.blueprint = path
				}
			}

			extInfos[idx] = info
		}(i, automationID)
	}
	wg.Wait()

	// Merge extended info back into items.
	for i, info := range extInfos {
		if info.description != "" {
			items[i]["description"] = info.description
		}
		if info.blueprint != "" {
			items[i]["blueprint"] = info.blueprint
		}
	}
	return nil
}

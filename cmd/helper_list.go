package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var helperListFlags *ListFlags

var helperListCmd = &cobra.Command{
	Use:     "list [type]",
	Short:   "List helper entities",
	Long:    `List all helper entities, optionally filtered by type.`,
	GroupID: helperGroupCommands,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runHelperList,
}

func init() {
	helperCmd.AddCommand(helperListCmd)
	helperListFlags = RegisterListFlags(helperListCmd, "entity_id")
}

func runHelperList(cmd *cobra.Command, args []string) error {
	var filterType string
	if len(args) > 0 {
		filterType = args[0]
	}

	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	entities, err := ws.EntityRegistryList()
	if err != nil {
		return err
	}

	// Helper domains
	helperDomains := map[string]bool{
		"input_boolean":  true,
		"input_number":   true,
		"input_text":     true,
		"input_select":   true,
		"input_datetime": true,
		"input_button":   true,
		"counter":        true,
		"timer":          true,
		"schedule":       true,
	}

	var result []map[string]interface{}
	for _, e := range entities {
		entity, ok := e.(map[string]interface{})
		if !ok {
			continue
		}

		entityID, _ := entity["entity_id"].(string)
		parts := strings.SplitN(entityID, ".", 2)
		if len(parts) < 2 {
			continue
		}

		domain := parts[0]
		if !helperDomains[domain] {
			continue
		}

		if filterType != "" && domain != filterType && domain != "input_"+filterType {
			continue
		}

		result = append(result, map[string]interface{}{
			"entity_id": entityID,
			"name":      entity["name"],
			"type":      domain,
		})
	}

	if helperListFlags.RenderCount(len(result), textMode) {
		return nil
	}
	result = helperListFlags.ApplyLimitMap(result)
	if helperListFlags.RenderBriefMap(result, textMode, "entity_id", "name") {
		return nil
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

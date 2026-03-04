package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var scriptListFlags *ListFlags

var scriptListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all scripts",
	Long:    `List all scripts in Home Assistant.`,
	GroupID: scriptGroupCommands,
	RunE:    runScriptList,
}

func init() {
	scriptCmd.AddCommand(scriptListCmd)
	scriptListFlags = RegisterListFlags(scriptListCmd, "entity_id")
}

func runScriptList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
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

	if scriptListFlags.RenderCount(len(result), textMode) {
		return nil
	}
	result = scriptListFlags.ApplyLimitMap(result)
	if scriptListFlags.RenderBriefMap(result, textMode, "entity_id", "alias") {
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
		output.PrintOutput(result, false, "")
	}
	return nil
}

package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var todoListsCmd = &cobra.Command{
	Use:   "lists",
	Short: "List to-do list entities",
	Long:  `List all to-do list entities registered in Home Assistant.`,
	Example: `  hab todo lists
  hab todo lists --json`,
	RunE: runTodoLists,
}

func init() {
	todoCmd.AddCommand(todoListsCmd)
}

func runTodoLists(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	states, err := restClient.GetStates()
	if err != nil {
		return err
	}

	var lists []interface{}
	for _, s := range states {
		state, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		entityID, _ := state["entity_id"].(string)
		if !strings.HasPrefix(entityID, "todo.") {
			continue
		}
		attrs, _ := state["attributes"].(map[string]interface{})
		name, _ := attrs["friendly_name"].(string)
		itemCount := state["state"] // HA stores item count as state
		lists = append(lists, map[string]interface{}{
			"entity_id":  entityID,
			"name":       name,
			"item_count": itemCount,
		})
	}

	if len(lists) == 0 {
		output.PrintOutput([]interface{}{}, textMode, "No to-do lists found.")
		return nil
	}

	output.PrintOutput(lists, textMode, "")
	return nil
}

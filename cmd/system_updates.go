package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var systemUpdatesCmd = &cobra.Command{
	Use:   "updates",
	Short: "Check for available updates",
	Long:  `Check for available updates to Home Assistant and add-ons.`,
	RunE:  runSystemUpdates,
}

func init() {
	systemCmd.AddCommand(systemUpdatesCmd)
}

func runSystemUpdates(cmd *cobra.Command, args []string) error {
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

	var updates []map[string]interface{}
	for _, s := range states {
		state, ok := s.(map[string]interface{})
		if !ok {
			continue
		}

		entityID, _ := state["entity_id"].(string)
		if !strings.HasPrefix(entityID, "update.") {
			continue
		}

		attrs, _ := state["attributes"].(map[string]interface{})
		title := entityID
		if t, ok := attrs["title"].(string); ok {
			title = t
		}

		updates = append(updates, map[string]interface{}{
			"entity_id":         entityID,
			"title":             title,
			"installed_version": attrs["installed_version"],
			"latest_version":    attrs["latest_version"],
			"update_available":  state["state"] == "on",
		})
	}

	output.PrintOutput(updates, textMode, "")
	return nil
}

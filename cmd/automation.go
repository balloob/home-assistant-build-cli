package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

const (
	automationGroupCommands    = "commands"
	automationGroupSubcommands = "subcommands"
)

var automationCmd = &cobra.Command{
	Use:     "automation",
	Short:   "Manage automations",
	Long:    `Create, update, delete, and trigger automations.`,
	GroupID: "automation",
}

// resolveAutomationConfigID converts an automation entity_id (e.g.
// "automation.good_night" or "good_night") to the internal config ID stored
// in attributes.id (e.g. "1682897162401"). This is required for the
// /api/config/automation/config/<id> REST endpoints which stopped accepting
// entity slugs in HA 2024.4+.
//
// If the state lookup fails (e.g. user provided a raw config ID directly),
// the input slug is returned as-is as a fallback so callers do not have to
// handle two code paths.
func resolveAutomationConfigID(restClient client.RestAPI, entityOrConfigID string) (string, error) {
	entityID := entityOrConfigID
	if !strings.HasPrefix(entityID, "automation.") {
		entityID = "automation." + entityID
	}

	state, err := restClient.GetState(entityID)
	if err != nil {
		// Fall back: assume the caller passed the raw internal ID directly.
		return strings.TrimPrefix(entityOrConfigID, "automation."), nil
	}

	attrs, ok := state["attributes"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("automation %s has no attributes", entityID)
	}

	configID, ok := attrs["id"].(string)
	if !ok || configID == "" {
		return "", fmt.Errorf("automation %s has no internal config ID in attributes", entityID)
	}

	return configID, nil
}

func init() {
	rootCmd.AddCommand(automationCmd)

	automationCmd.AddGroup(
		&cobra.Group{ID: automationGroupCommands, Title: "Commands:"},
		&cobra.Group{ID: automationGroupSubcommands, Title: "Subcommands:"},
	)
}

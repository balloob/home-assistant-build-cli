package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

const (
	automationGroupCommands    = groupCommands
	automationGroupSubcommands = groupSubcommands
)

var automationCmd = &cobra.Command{
	Use:     "automation",
	Short:   "Manage automations",
	Long:    `Create, update, delete, and trigger automations.`,
	GroupID: "automation",
}

// resolveAutomationConfigID converts an automation entity_id (e.g.
// "automation.good_night" or "good_night") to the internal config ID stored
// in attributes.id. Delegates to the shared resolveStateConfigID helper.
func resolveAutomationConfigID(restClient client.RestAPI, entityOrConfigID string) (string, error) {
	return resolveStateConfigID(restClient, "automation", entityOrConfigID)
}

func init() {
	rootCmd.AddCommand(automationCmd)

	automationCmd.AddGroup(
		&cobra.Group{ID: automationGroupCommands, Title: "Commands:"},
		&cobra.Group{ID: automationGroupSubcommands, Title: "Subcommands:"},
	)

	RegisterConfigResourceCRUD(ConfigResourceConfig{
		ParentCmd:    automationCmd,
		ResourceName: "automation",
		APIPrefix:    "config/automation/config/",
		IDFlagName:   "automation",
		ResolveID:    resolveAutomationConfigID,
		GroupID:      automationGroupCommands,
		CreateExample: `  hab automation create my_lights -d '{"alias":"Evening Lights","triggers":[{"trigger":"sun","event":"sunset"}],"actions":[{"action":"light.turn_on","target":{"entity_id":"light.living_room"}}]}'
  hab automation create morning_routine -f automation.yaml`,
	})

	RegisterSubComponentCRUD(SubComponentConfig{
		ParentCmd:       automationCmd,
		ParentName:      "automation",
		ComponentName:   "trigger",
		ComponentPlural: "triggers",
		ConfigKeys:      []string{"triggers", "trigger"},
		DefaultKey:      "triggers",
		APIBasePath:     "config/automation/config/",
		ResolveID:       resolveAutomationConfigID,
		ParentFlagName:  "automation",
		GroupID:         automationGroupSubcommands,
	})

	RegisterSubComponentCRUD(SubComponentConfig{
		ParentCmd:       automationCmd,
		ParentName:      "automation",
		ComponentName:   "condition",
		ComponentPlural: "conditions",
		ConfigKeys:      []string{"conditions", "condition"},
		DefaultKey:      "conditions",
		APIBasePath:     "config/automation/config/",
		ResolveID:       resolveAutomationConfigID,
		ParentFlagName:  "automation",
		GroupID:         automationGroupSubcommands,
	})

	RegisterSubComponentCRUD(SubComponentConfig{
		ParentCmd:       automationCmd,
		ParentName:      "automation",
		ComponentName:   "action",
		ComponentPlural: "actions",
		ConfigKeys:      []string{"actions", "action"},
		DefaultKey:      "actions",
		APIBasePath:     "config/automation/config/",
		ResolveID:       resolveAutomationConfigID,
		ParentFlagName:  "automation",
		GroupID:         automationGroupSubcommands,
	})
}

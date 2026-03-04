package cmd

import (
	"github.com/spf13/cobra"
)

const (
	dashboardGroupCommands    = "commands"
	dashboardGroupSubcommands = "subcommands"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Manage dashboards",
	Long: `Create, update, and delete dashboards.

If you are new to creating Home Assistant dashboards, you MUST run 'hab dashboard guide' first.`,
	GroupID: "dashboard",
}

func init() {
	rootCmd.AddCommand(dashboardCmd)

	dashboardCmd.AddGroup(
		&cobra.Group{ID: dashboardGroupCommands, Title: "Commands:"},
		&cobra.Group{ID: dashboardGroupSubcommands, Title: "Subcommands:"},
	)

	// ── Dashboard sub-resource CRUD (view, badge, section, card) ──

	RegisterDashboardResourceCRUD(DashboardResourceConfig{
		ResourceName: "view",
		ParentCmd:    dashboardCmd,
		GroupID:      dashboardGroupSubcommands,
		ShortDesc:    "Manage dashboard views",
		LongDesc:     "Create, update, list, and delete views in a dashboard.",
		PathFromConfig: []string{"views"},
		GetUsesFlags:   true,
		CreateFlags: []DashboardResourceFlag{
			{Name: "title", Usage: "View title", ConfigKey: "title"},
			{Name: "icon", Usage: "View icon (e.g., mdi:home)", ConfigKey: "icon"},
			{Name: "path", Usage: "View path (URL slug)", ConfigKey: "path"},
		},
		CreateRequiredFields: []string{"title"},
		UpdateFlags: []DashboardResourceFlag{
			{Name: "title", Usage: "View title", ConfigKey: "title"},
			{Name: "icon", Usage: "View icon (e.g., mdi:home)", ConfigKey: "icon"},
			{Name: "path", Usage: "View path (URL slug)", ConfigKey: "path"},
		},
	})

	RegisterDashboardResourceCRUD(DashboardResourceConfig{
		ResourceName: "badge",
		ParentCmd:    dashboardCmd,
		GroupID:      dashboardGroupSubcommands,
		ShortDesc:    "Manage view badges",
		LongDesc:     "Create, update, list, and delete badges in a dashboard view.",
		PathFromConfig:  []string{"views", "badges"},
		ItemCanBeString: true,
		CreateFlags: []DashboardResourceFlag{
			{Name: "entity", Usage: "Entity ID for simple badge", ConfigKey: "entity"},
			{Name: "type", Usage: "Badge type (e.g., entity)", ConfigKey: "type"},
		},
		UpdateFlags: []DashboardResourceFlag{
			{Name: "entity", Usage: "Entity ID for simple badge", ConfigKey: "entity"},
		},
	})

	RegisterDashboardResourceCRUD(DashboardResourceConfig{
		ResourceName: "section",
		ParentCmd:    dashboardCmd,
		GroupID:      dashboardGroupSubcommands,
		ShortDesc:    "Manage view sections",
		LongDesc:     "Create, update, list, and delete sections in a dashboard view.",
		PathFromConfig: []string{"views", "sections"},
		GetUsesFlags:   true,
		CreateFlags: []DashboardResourceFlag{
			{Name: "title", Usage: "Section title", ConfigKey: "title"},
			{Name: "type", Usage: "Section type (e.g., grid)", ConfigKey: "type"},
		},
		CreateDefaults: map[string]interface{}{
			"cards": []interface{}{},
		},
		UpdateFlags: []DashboardResourceFlag{
			{Name: "title", Usage: "Section title", ConfigKey: "title"},
			{Name: "type", Usage: "Section type", ConfigKey: "type"},
		},
	})

	RegisterDashboardResourceCRUD(DashboardResourceConfig{
		ResourceName: "card",
		ParentCmd:    dashboardCmd,
		GroupID:      dashboardGroupSubcommands,
		ShortDesc:    "Manage dashboard cards",
		LongDesc:     "Create, update, list, and delete cards in a dashboard view or section.",
		PathFromConfig:        []string{"views", "sections", "cards"},
		HasSectionFlag:        true,
		GetUsesFlags:          true,
		CreateAutoScaffold:    true,
		CreateViewArgOptional: true,
		CreateLongDesc: `Create a new card in a dashboard view or section.

If view_index is not specified, uses the last view. If no views exist, creates one.
If section is not specified, uses the last section. If no sections exist, creates one.
If type is not specified, defaults to "tile".`,
		CreateFlags: []DashboardResourceFlag{
			{Name: "type", Usage: "Card type (e.g., entities, button, markdown)", ConfigKey: "type"},
			{Name: "entity", Usage: "Entity ID (for simple entity cards)", ConfigKey: "entity"},
			{Name: "name", Usage: "Card name/title", ConfigKey: "name"},
		},
		CreateDefaults: map[string]interface{}{
			"type": "tile",
		},
		UpdateLongDesc: "Update a card in a section by index.\n\nIf section is not specified, uses the last section.",
		UpdateFlags: []DashboardResourceFlag{
			{Name: "type", Usage: "Card type", ConfigKey: "type"},
			{Name: "entity", Usage: "Entity ID", ConfigKey: "entity"},
		},
		DeleteLongDesc: "Delete a card from a section by index.\n\nIf section is not specified, uses the last section.",
		ListLongDesc:   "List all cards in a dashboard section.\n\nIf section is not specified, uses the last section.",
	})
}

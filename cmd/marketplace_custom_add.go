package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceCustomAddCategory string

var marketplaceCustomAddCmd = &cobra.Command{
	Use:   "add <github_repository>",
	Short: "Add a custom Marketplace repository",
	Long: `Add a GitHub repository to the Marketplace. Give it as owner/repo or as a
GitHub URL. Run 'hab marketplace custom detect' to find its category.`,
	Example: `  hab marketplace custom add custom-cards/button-card --category plugin`,
	Args:    cobra.ExactArgs(1),
	RunE:    runMarketplaceCustomAdd,
}

func init() {
	marketplaceCustomCmd.AddCommand(marketplaceCustomAddCmd)
	marketplaceCustomAddCmd.Flags().StringVar(&marketplaceCustomAddCategory, "category", "", "Category: integration, plugin, theme, or template")
	marketplaceCustomAddCmd.MarkFlagRequired("category")
	mergeSchemaAnnotation(marketplaceCustomAddCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args", "flags"},
	})
}

func runMarketplaceCustomAdd(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if _, err := ws.SendCommand("marketplace/repositories/add", map[string]interface{}{
		"repository": args[0],
		"category":   marketplaceCustomAddCategory,
	}); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Repository '%s' added.", args[0]), output.EnvelopeContext{
		Operation:             "add",
		ResourceType:          "marketplace_repository",
		NextSuggestedCommands: []string{fmt.Sprintf("hab marketplace install %s", args[0])},
	})
	return nil
}

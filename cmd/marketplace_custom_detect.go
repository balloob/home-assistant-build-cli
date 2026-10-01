package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceCustomDetectCmd = &cobra.Command{
	Use:   "detect <github_repository>",
	Short: "Detect the categories of a GitHub repository",
	Long: `Show the Marketplace categories a GitHub repository holds content for,
before you add it. Give the repository as owner/repo or as a GitHub URL.`,
	Example: `  hab marketplace custom detect custom-cards/button-card`,
	Args:    cobra.ExactArgs(1),
	RunE:    runMarketplaceCustomDetect,
}

func init() {
	marketplaceCustomCmd.AddCommand(marketplaceCustomDetectCmd)
	mergeSchemaAnnotation(marketplaceCustomDetectCmd, SchemaAnnotation{
		SideEffect:   "read",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args"},
	})
}

func runMarketplaceCustomDetect(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("marketplace/repositories/detect", map[string]interface{}{
		"repository": args[0],
	})
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

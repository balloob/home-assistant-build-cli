package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceCustomRemoveCmd = &cobra.Command{
	Use:   "remove <repository>",
	Short: "Remove a custom Marketplace repository",
	Long: `Remove a custom repository from the Marketplace list. Uninstall it first
if it is installed.`,
	Args: cobra.ExactArgs(1),
	RunE: runMarketplaceCustomRemove,
}

func init() {
	marketplaceCustomCmd.AddCommand(marketplaceCustomRemoveCmd)
	// Only the list entry goes; Home Assistant refuses while files are installed
	mergeSchemaAnnotation(marketplaceCustomRemoveCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args"},
	})
}

func runMarketplaceCustomRemove(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	repositoryID, err := resolveMarketplaceRepository(ws, args[0])
	if err != nil {
		return err
	}

	if _, err := ws.SendCommand("marketplace/repositories/remove", map[string]interface{}{
		"repository": repositoryID,
	}); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Repository '%s' removed.", args[0]), output.EnvelopeContext{Operation: "remove", ResourceType: "marketplace_repository"})
	return nil
}

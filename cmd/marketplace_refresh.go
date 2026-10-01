package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceRefreshCmd = &cobra.Command{
	Use:   "refresh <repository>",
	Short: "Refresh a Marketplace repository",
	Long:  `Fetch the latest information of a repository from GitHub, such as new releases.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMarketplaceRefresh,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceRefreshCmd)
	mergeSchemaAnnotation(marketplaceRefreshCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args"},
	})
}

func runMarketplaceRefresh(cmd *cobra.Command, args []string) error {
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

	if _, err := ws.SendCommand("marketplace/repository/refresh", map[string]interface{}{
		"repository": repositoryID,
	}); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Repository '%s' refreshed.", args[0]), output.EnvelopeContext{Operation: "refresh", ResourceType: "marketplace_repository"})
	return nil
}

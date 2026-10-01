package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceIgnoreCmd = &cobra.Command{
	Use:   "ignore <repository>",
	Short: "Ignore a removed Marketplace repository",
	Long: `Stop reporting that the Marketplace catalog removed a repository.
It no longer shows in 'hab marketplace removed'.`,
	Args: cobra.ExactArgs(1),
	RunE: runMarketplaceIgnore,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceIgnoreCmd)
	mergeSchemaAnnotation(marketplaceIgnoreCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args"},
	})
}

func runMarketplaceIgnore(cmd *cobra.Command, args []string) error {
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

	if _, err := ws.SendCommand("marketplace/repository/ignore", map[string]interface{}{
		"repository": repositoryID,
	}); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Repository '%s' ignored.", args[0]), output.EnvelopeContext{Operation: "ignore", ResourceType: "marketplace_repository"})
	return nil
}

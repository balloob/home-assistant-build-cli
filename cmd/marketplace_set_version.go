package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceSetVersionCmd = &cobra.Command{
	Use:   "set-version <repository> <version>",
	Short: "Select the version a Marketplace repository updates to",
	Long: `Select the release tag or branch that the update entity of a repository
offers. Give the default branch to go back to the latest release.

This does not install anything. Run 'hab marketplace install' to install it.`,
	Example: `  hab marketplace set-version piitaya/lovelace-mushroom v4.0.0
  hab marketplace set-version piitaya/lovelace-mushroom main`,
	Args: cobra.ExactArgs(2),
	RunE: runMarketplaceSetVersion,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceSetVersionCmd)
	mergeSchemaAnnotation(marketplaceSetVersionCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args"},
	})
}

func runMarketplaceSetVersion(cmd *cobra.Command, args []string) error {
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

	if _, err := ws.SendCommand("marketplace/repository/version", map[string]interface{}{
		"repository": repositoryID,
		"version":    args[1],
	}); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Repository '%s' set to version %s.", args[0], args[1]), output.EnvelopeContext{Operation: "set-version", ResourceType: "marketplace_repository"})
	return nil
}

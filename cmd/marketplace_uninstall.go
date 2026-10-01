package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceUninstallForce bool

var marketplaceUninstallCmd = &cobra.Command{
	Use:   "uninstall <repository>",
	Short: "Uninstall a Marketplace repository",
	Long: `Remove the files of an installed Marketplace repository.

Home Assistant refuses to uninstall an integration that still has config
entries. Delete them first ('hab integration list --domain <domain>').`,
	Args: cobra.ExactArgs(1),
	RunE: runMarketplaceUninstall,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceUninstallCmd)
	marketplaceUninstallCmd.Flags().BoolVarP(&marketplaceUninstallForce, "force", "f", false, "Skip confirmation")
	mergeSchemaAnnotation(marketplaceUninstallCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args", "flags"},
	})
}

func runMarketplaceUninstall(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	if err := confirmAction(marketplaceUninstallForce, fmt.Sprintf("Uninstall repository %s?", args[0]), "uninstall repository"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	repositoryID, err := resolveMarketplaceRepository(ws, args[0])
	if err != nil {
		return err
	}

	if _, err := ws.SendCommand("marketplace/repository/uninstall", map[string]interface{}{
		"repository": repositoryID,
	}); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Repository '%s' uninstalled.", args[0]), output.EnvelopeContext{Operation: "uninstall", ResourceType: "marketplace_repository"})
	return nil
}

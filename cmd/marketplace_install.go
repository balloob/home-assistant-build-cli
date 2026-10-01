package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	marketplaceInstallVersion        string
	marketplaceInstallReplaceBuiltIn bool
)

var marketplaceInstallCmd = &cobra.Command{
	Use:   "install <repository>",
	Short: "Install a Marketplace repository",
	Long: `Download and install a Marketplace repository, or install another
version of an installed one.

An integration that has the same domain as a built-in integration replaces it.
Home Assistant refuses that unless --confirm-replace-built-in is given.

Some installs need a restart. Home Assistant then raises a repair issue
('hab repairs list').`,
	Example: `  hab marketplace install piitaya/lovelace-mushroom
  hab marketplace install hacs/integration-blueprint --version v1.2.0`,
	Args: cobra.ExactArgs(1),
	RunE: runMarketplaceInstall,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceInstallCmd)
	marketplaceInstallCmd.Flags().StringVar(&marketplaceInstallVersion, "version", "", "Release tag or branch to install (default: the available version)")
	marketplaceInstallCmd.Flags().BoolVar(&marketplaceInstallReplaceBuiltIn, "confirm-replace-built-in", false, "Allow the integration to replace a built-in integration with the same domain")
	mergeSchemaAnnotation(marketplaceInstallCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args", "flags"},
	})
}

func runMarketplaceInstall(cmd *cobra.Command, args []string) error {
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

	params := map[string]interface{}{"repository": repositoryID}
	if marketplaceInstallVersion != "" {
		params["version"] = marketplaceInstallVersion
	}
	if marketplaceInstallReplaceBuiltIn {
		params["confirm_replace_built_in"] = true
	}
	if _, err := ws.SendCommand("marketplace/repository/install", params); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Repository '%s' installed.", args[0]), output.EnvelopeContext{
		Operation:            "install",
		ResourceType:         "marketplace_repository",
		VerificationCommands: []string{fmt.Sprintf("hab marketplace get %s --json", args[0]), "hab repairs list --json"},
	})
	return nil
}

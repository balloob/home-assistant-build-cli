package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceAcceptWarningCmd = &cobra.Command{
	Use:   "accept-warning",
	Short: "Accept the Marketplace warning",
	Long: `Store that the current user read and accepted the Marketplace warning.

Marketplace content comes from the community. Home Assistant does not review
it, and it runs with full access to your Home Assistant. The Marketplace
refuses installs and version changes until the user accepts this.

Ask the user before you run this command for them.`,
	RunE: runMarketplaceAcceptWarning,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceAcceptWarningCmd)
	mergeSchemaAnnotation(marketplaceAcceptWarningCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace",
		InputSources: []string{"flags"},
	})
}

func runMarketplaceAcceptWarning(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if _, err := ws.SendCommand("marketplace/warning/accept", nil); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, "Marketplace warning accepted.", output.EnvelopeContext{Operation: "accept-warning", ResourceType: "marketplace"})
	return nil
}

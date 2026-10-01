package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceCriticalAcknowledgeCmd = &cobra.Command{
	Use:   "acknowledge <github_repository>",
	Short: "Acknowledge a critical Marketplace repository",
	Long:  `Mark a critical repository as acknowledged. Give the repository as owner/repo, as 'hab marketplace critical list' shows it.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMarketplaceCriticalAcknowledge,
}

func init() {
	marketplaceCriticalCmd.AddCommand(marketplaceCriticalAcknowledgeCmd)
	mergeSchemaAnnotation(marketplaceCriticalAcknowledgeCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args"},
	})
}

func runMarketplaceCriticalAcknowledge(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("marketplace/critical/acknowledge", map[string]interface{}{
		"repository": args[0],
	})
	if err != nil {
		return err
	}

	output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("Critical repository '%s' acknowledged.", args[0]), output.EnvelopeContext{Operation: "acknowledge", ResourceType: "marketplace_repository"})
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceCriticalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List critical Marketplace repositories",
	Long:  `List the critical repositories that were installed here, with the reason and whether a user acknowledged them.`,
	RunE:  runMarketplaceCriticalList,
}

func init() {
	marketplaceCriticalCmd.AddCommand(marketplaceCriticalListCmd)
	mergeSchemaAnnotation(marketplaceCriticalListCmd, SchemaAnnotation{
		SideEffect:   "read",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
	})
}

func runMarketplaceCriticalList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("marketplace/critical/list", nil)
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

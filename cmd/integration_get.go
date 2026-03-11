package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var integrationGetCmd = &cobra.Command{
	Use:   "get <entry_id>",
	Short: "Get integration details",
	Long:  `Get full details for a specific integration by its config entry ID.`,
	Example: `  hab integration get abc123def456
  hab integration get abc123def456 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runIntegrationGet,
}

func init() {
	integrationCmd.AddCommand(integrationGetCmd)
}

func runIntegrationGet(cmd *cobra.Command, args []string) error {
	entryID := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	entry, err := ws.ConfigEntryGet(entryID)
	if err != nil {
		return err
	}

	output.PrintOutput(entry, textMode, "")
	return nil
}

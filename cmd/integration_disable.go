package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var integrationDisableCmd = &cobra.Command{
	Use:   "disable <entry_id>",
	Short: "Disable an integration",
	Long:  `Disable a Home Assistant integration. It can be re-enabled with 'integration enable'.`,
	Example: `  hab integration disable abc123def456`,
	Args: cobra.ExactArgs(1),
	RunE: runIntegrationDisable,
}

func init() {
	integrationCmd.AddCommand(integrationDisableCmd)
}

func runIntegrationDisable(cmd *cobra.Command, args []string) error {
	entryID := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.ConfigEntrySetDisabled(entryID, "user")
	if err != nil {
		return err
	}

	if requireRestart, ok := result["require_restart"].(bool); ok && requireRestart {
		output.PrintSuccess(result, textMode, "Integration disabled. A Home Assistant restart is required.")
	} else {
		output.PrintSuccess(result, textMode, "Integration disabled.")
	}
	return nil
}

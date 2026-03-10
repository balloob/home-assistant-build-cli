package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var integrationEnableCmd = &cobra.Command{
	Use:   "enable <entry_id>",
	Short: "Enable an integration",
	Long:  `Enable a previously disabled Home Assistant integration.`,
	Example: `  hab integration enable abc123def456`,
	Args: cobra.ExactArgs(1),
	RunE: runIntegrationEnable,
}

func init() {
	integrationCmd.AddCommand(integrationEnableCmd)
}

func runIntegrationEnable(cmd *cobra.Command, args []string) error {
	entryID := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.ConfigEntrySetDisabled(entryID, nil)
	if err != nil {
		return err
	}

	if requireRestart, ok := result["require_restart"].(bool); ok && requireRestart {
		output.PrintSuccess(result, textMode, "Integration enabled. A Home Assistant restart is required.")
	} else {
		output.PrintSuccess(result, textMode, "Integration enabled.")
	}
	return nil
}

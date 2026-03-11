package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var integrationReloadCmd = &cobra.Command{
	Use:   "reload <entry_id>",
	Short: "Reload an integration",
	Long:  `Reload a Home Assistant integration without restarting Home Assistant.`,
	Example: `  hab integration reload abc123def456`,
	Args: cobra.ExactArgs(1),
	RunE: runIntegrationReload,
}

func init() {
	integrationCmd.AddCommand(integrationReloadCmd)
}

func runIntegrationReload(cmd *cobra.Command, args []string) error {
	entryID := args[0]
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	result, err := restClient.ConfigEntryReload(entryID)
	if err != nil {
		return err
	}

	if requireRestart, ok := result["require_restart"].(bool); ok && requireRestart {
		output.PrintSuccess(result, textMode, "Integration reloaded. A Home Assistant restart is required to complete.")
	} else {
		output.PrintSuccess(result, textMode, "Integration reloaded.")
	}
	return nil
}

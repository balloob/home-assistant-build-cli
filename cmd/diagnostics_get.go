package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var diagnosticsGetCmd = &cobra.Command{
	Use:   "get <domain>",
	Short: "Get diagnostics handler details",
	Long:  `Get diagnostics handler details for an integration domain.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDiagnosticsGet,
}

func init() {
	diagnosticsCmd.AddCommand(diagnosticsGetCmd)
}

func runDiagnosticsGet(cmd *cobra.Command, args []string) error {
	domain := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.DiagnosticsGet(domain)
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

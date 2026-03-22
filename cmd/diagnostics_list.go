package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var diagnosticsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List diagnostics handlers",
	Long:  `List diagnostics handlers exposed by integrations.`,
	RunE:  runDiagnosticsList,
}

func init() {
	diagnosticsCmd.AddCommand(diagnosticsListCmd)
}

func runDiagnosticsList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.DiagnosticsList()
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

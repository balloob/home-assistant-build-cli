package cmd

import "github.com/spf13/cobra"

var diagnosticsCmd = &cobra.Command{
	Use:   "diagnostics",
	Short: "Manage diagnostics handlers",
	Long: `Inspect available diagnostics handlers for integrations.

For workflow guidance, run 'hab guide operations'.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(diagnosticsCmd)
}

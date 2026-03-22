package cmd

import "github.com/spf13/cobra"

var diagnosticsCmd = &cobra.Command{
	Use:     "diagnostics",
	Short:   "Manage diagnostics handlers",
	Long:    `Inspect available diagnostics handlers for integrations.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(diagnosticsCmd)
}

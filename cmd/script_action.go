package cmd

import (
	"github.com/spf13/cobra"
)

var scriptActionCmd = &cobra.Command{
	Use:   "script-action",
	Short: "Manage script actions",
	Long:  `Create, update, list, and delete actions in a script's sequence.`,
}

func init() {
	rootCmd.AddCommand(scriptActionCmd)
}

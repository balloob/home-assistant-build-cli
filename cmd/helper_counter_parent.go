package cmd

import (
	"github.com/spf13/cobra"
)

var helperCounterParentCmd = &cobra.Command{
	Use:   "helper-counter",
	Short: "Manage counter helpers",
	Long:  `Create, list, and delete counter helpers.`,
}

func init() {
	rootCmd.AddCommand(helperCounterParentCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
)

var helperMinMaxParentCmd = &cobra.Command{
	Use:   "helper-min-max",
	Short: "Manage min/max sensor helpers",
	Long:  `Create, list, and delete min/max sensor helpers that aggregate values from multiple sensors.`,
}

func init() {
	rootCmd.AddCommand(helperMinMaxParentCmd)
}

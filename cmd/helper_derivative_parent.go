package cmd

import (
	"github.com/spf13/cobra"
)

var helperDerivativeParentCmd = &cobra.Command{
	Use:   "helper-derivative",
	Short: "Manage derivative sensor helpers",
	Long:  `Create, list, and delete derivative sensor helpers that calculate the rate of change of a source sensor.`,
}

func init() {
	rootCmd.AddCommand(helperDerivativeParentCmd)
}

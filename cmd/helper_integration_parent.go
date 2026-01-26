package cmd

import (
	"github.com/spf13/cobra"
)

var helperIntegrationParentCmd = &cobra.Command{
	Use:     "integration",
	Short:   "Manage integration (integral) sensor helpers",
	Long:    `Create, list, and delete integration sensor helpers that calculate the Riemann sum (integral) of a source sensor.`,
	GroupID: helperGroupSubcommands,
}

func init() {
	helperCmd.AddCommand(helperIntegrationParentCmd)
}

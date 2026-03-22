package cmd

import "github.com/spf13/cobra"

var energyCmd = &cobra.Command{
	Use:     "energy",
	Short:   "Manage energy dashboard settings",
	Long:    `Inspect and manage Home Assistant energy dashboard preferences and metadata.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(energyCmd)
}

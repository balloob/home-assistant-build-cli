package cmd

import "github.com/spf13/cobra"

var energyPrefsCmd = &cobra.Command{
	Use:   "prefs",
	Short: "Manage energy preferences",
	Long:  `Get and update energy dashboard preferences.`,
}

func init() {
	energyCmd.AddCommand(energyPrefsCmd)
}

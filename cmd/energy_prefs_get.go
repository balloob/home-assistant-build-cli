package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var energyPrefsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get energy preferences",
	Long:  `Get current energy dashboard preferences.`,
	RunE:  runEnergyPrefsGet,
}

func init() {
	energyPrefsCmd.AddCommand(energyPrefsGetCmd)
}

func runEnergyPrefsGet(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.EnergyGetPrefs()
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

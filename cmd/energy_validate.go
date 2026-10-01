package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var energyValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate energy preferences",
	Long:  `Validate the saved energy dashboard preferences.`,
	RunE:  runEnergyValidate,
}

func init() {
	energyCmd.AddCommand(energyValidateCmd)
}

func runEnergyValidate(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.EnergyValidate()
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

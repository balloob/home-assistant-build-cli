package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var energyInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get energy dashboard info",
	Long:  `Get energy dashboard metadata, such as cost sensor mappings and supported forecast domains.`,
	RunE:  runEnergyInfo,
}

func init() {
	energyCmd.AddCommand(energyInfoCmd)
}

func runEnergyInfo(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.EnergyInfo()
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

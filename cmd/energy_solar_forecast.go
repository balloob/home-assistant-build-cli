package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var energySolarForecastCmd = &cobra.Command{
	Use:   "solar-forecast",
	Short: "Get solar forecast",
	Long:  `Get solar forecast data used by the energy dashboard.`,
	RunE:  runEnergySolarForecast,
}

func init() {
	energyCmd.AddCommand(energySolarForecastCmd)
}

func runEnergySolarForecast(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.EnergySolarForecast()
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

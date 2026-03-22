package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var energyPrefsSetInput InputFlags

var energyPrefsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set energy preferences",
	Long:  `Set energy dashboard preferences from JSON or YAML input.`,
	RunE:  runEnergyPrefsSet,
}

func init() {
	energyPrefsCmd.AddCommand(energyPrefsSetCmd)
	energyPrefsSetInput.Register(energyPrefsSetCmd)
}

func runEnergyPrefsSet(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	if energyPrefsSetInput.Data == "" && energyPrefsSetInput.File == "" {
		return fmt.Errorf("energy preferences data is required (use --data or --file)")
	}

	params, err := energyPrefsSetInput.Parse()
	if err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.EnergySavePrefs(params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, "Energy preferences updated.")
	return nil
}

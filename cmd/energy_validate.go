package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var energyValidateInput InputFlags

var energyValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate energy preferences",
	Long:  `Validate energy dashboard preferences. Optionally pass preferences via --data/--file for server-side validation without saving.`,
	RunE:  runEnergyValidate,
}

func init() {
	energyCmd.AddCommand(energyValidateCmd)
	energyValidateInput.Register(energyValidateCmd)
}

func runEnergyValidate(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	params := map[string]interface{}{}
	if energyValidateInput.Data != "" || energyValidateInput.File != "" {
		parsed, err := energyValidateInput.Parse()
		if err != nil {
			return err
		}
		params = parsed
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.EnergyValidate(params)
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

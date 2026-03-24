package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var esphomeValidateStructured bool

var esphomeValidateCmd = &cobra.Command{
	Use:   "validate [configuration]",
	Short: "Validate ESPHome device configuration",
	Long: `Validate an ESPHome device configuration.

By default this streams the native ESPHome validator output. Use --structured
to return a single parsed result with richer structured errors in JSON mode.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runESPHomeValidate,
}

func init() {
	esphomeCmd.AddCommand(esphomeValidateCmd)
	esphomeValidateCmd.Flags().BoolVar(&esphomeValidateStructured, "structured", false, "Return a parsed validation result instead of streaming raw output")
}

func runESPHomeValidate(cmd *cobra.Command, args []string) error {
	configuration, err := resolveESPHomeConfiguration(args, 0)
	if err != nil {
		return err
	}
	textMode := getTextMode()

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	if !esphomeValidateStructured {
		return streamToOutput(esClient, "/validate", map[string]any{
			"type":          "spawn",
			"configuration": configuration,
		}, textMode)
	}

	parsed, err := esClient.GetJSONConfig(configuration)
	if err != nil {
		return err
	}

	if textMode {
		output.PrintSuccess(nil, true, fmt.Sprintf("Configuration %s is valid.", configuration))
		return nil
	}

	output.PrintOutput(map[string]any{
		"configuration": configuration,
		"parsed":        parsed,
	}, false, "")
	return nil
}

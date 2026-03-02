package cmd

import (
	"github.com/spf13/cobra"
)

var esphomeValidateCmd = &cobra.Command{
	Use:   "validate <configuration>",
	Short: "Validate ESPHome device configuration",
	Long: `Validate an ESPHome device configuration without compiling.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Output is streamed in real-time.`,
	Args: cobra.ExactArgs(1),
	RunE: runESPHomeValidate,
}

func init() {
	esphomeCmd.AddCommand(esphomeValidateCmd)
}

func runESPHomeValidate(cmd *cobra.Command, args []string) error {
	configuration := args[0]
	textMode := getTextMode()

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	spawnMsg := map[string]interface{}{
		"type":          "spawn",
		"configuration": configuration,
	}

	return streamToOutput(esClient, "/validate", spawnMsg, textMode)
}

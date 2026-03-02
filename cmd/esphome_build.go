package cmd

import (
	"github.com/spf13/cobra"
)

var esphomeBuildCmd = &cobra.Command{
	Use:   "build <configuration>",
	Short: "Compile ESPHome firmware",
	Long: `Compile firmware for an ESPHome device configuration.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Output is streamed in real-time as the build progresses.`,
	Args: cobra.ExactArgs(1),
	RunE: runESPHomeBuild,
}

func init() {
	esphomeCmd.AddCommand(esphomeBuildCmd)
}

func runESPHomeBuild(cmd *cobra.Command, args []string) error {
	configuration := args[0]
	textMode := getTextMode()

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	spawnMsg := map[string]interface{}{
		"type":          "spawn",
		"configuration": configuration,
		"only_generate": false,
	}

	return streamToOutput(esClient, "/compile", spawnMsg, textMode)
}

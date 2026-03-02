package cmd

import (
	"github.com/spf13/cobra"
)

var esphomeRunPort string

var esphomeRunCmd = &cobra.Command{
	Use:   "run <configuration>",
	Short: "Build and upload firmware to ESPHome device",
	Long: `Compile and flash firmware to an ESPHome device in one step.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Use --port to specify the connection method: "OTA" for network (default),
or a serial port path like "/dev/ttyUSB0".

This is equivalent to running 'esphome build' followed by 'esphome upload'.
Output is streamed in real-time.`,
	Args: cobra.ExactArgs(1),
	RunE: runESPHomeRun,
}

func init() {
	esphomeCmd.AddCommand(esphomeRunCmd)
	esphomeRunCmd.Flags().StringVar(&esphomeRunPort, "port", "OTA", "Connection port: OTA (network) or serial port path")
}

func runESPHomeRun(cmd *cobra.Command, args []string) error {
	configuration := args[0]
	textMode := getTextMode()

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	spawnMsg := map[string]interface{}{
		"type":          "spawn",
		"configuration": configuration,
		"port":          esphomeRunPort,
	}

	return streamToOutput(esClient, "/run", spawnMsg, textMode)
}

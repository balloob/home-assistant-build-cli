package cmd

import (
	"github.com/spf13/cobra"
)

var esphomeLogsPort string

var esphomeLogsCmd = &cobra.Command{
	Use:   "logs <configuration>",
	Short: "Stream live device logs",
	Long: `Stream live logs from an ESPHome device.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Use --port to specify the connection method: "OTA" for network (default),
or a serial port path like "/dev/ttyUSB0".

Logs are streamed in real-time until interrupted (Ctrl+C).`,
	Args: cobra.ExactArgs(1),
	RunE: runESPHomeLogs,
}

func init() {
	esphomeCmd.AddCommand(esphomeLogsCmd)
	esphomeLogsCmd.Flags().StringVar(&esphomeLogsPort, "port", "OTA", "Connection port: OTA (network) or serial port path")
}

func runESPHomeLogs(cmd *cobra.Command, args []string) error {
	configuration := args[0]
	textMode := getTextMode()

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	spawnMsg := map[string]interface{}{
		"type":          "spawn",
		"configuration": configuration,
		"port":          esphomeLogsPort,
	}

	return streamToOutput(esClient, "/logs", spawnMsg, textMode)
}

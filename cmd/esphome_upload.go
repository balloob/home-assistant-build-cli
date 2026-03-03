package cmd

import (
	"github.com/spf13/cobra"
)

var esphomeUploadPort string

var esphomeUploadCmd = &cobra.Command{
	Use:   "upload <configuration>",
	Short: "Upload firmware to ESPHome device",
	Long: `Flash compiled firmware to an ESPHome device.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Use --port to specify the connection method: "OTA" for network (default),
or a serial port path like "/dev/ttyUSB0".

The firmware must have been compiled first using 'esphome build'.
Output is streamed in real-time.`,
	Args: cobra.ExactArgs(1),
	RunE: runESPHomeUpload,
}

func init() {
	esphomeCmd.AddCommand(esphomeUploadCmd)
	esphomeUploadCmd.Flags().StringVar(&esphomeUploadPort, "port", "OTA", "Connection port: OTA (network) or serial port path")
}

func runESPHomeUpload(cmd *cobra.Command, args []string) error {
	configuration := args[0]
	textMode := getTextMode()

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	spawnMsg := map[string]interface{}{
		"type":          "spawn",
		"configuration": configuration,
		"port":          esphomeUploadPort,
	}

	return streamToOutput(esClient, "/upload", spawnMsg, textMode)
}

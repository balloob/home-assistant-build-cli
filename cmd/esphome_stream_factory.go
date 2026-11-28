package cmd

import (
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// ESPHome streaming command factory
//
// Generates subcommands for ESPHome streaming operations (build, validate,
// upload, run, logs) from a declarative configuration.
// ---------------------------------------------------------------------------

// ESPHomeStreamConfig holds all configuration needed to generate an ESPHome
// streaming subcommand.
type ESPHomeStreamConfig struct {
	Use       string // cobra command Use field, e.g. "build <configuration>"
	Short     string // one-line description
	Long      string // detailed description
	Operation string // client.StreamCommand operation, e.g. "build"
	HasPort   bool   // whether to add --port flag
}

// RegisterESPHomeStream generates and registers a streaming subcommand on
// esphomeCmd.
func RegisterESPHomeStream(cfg ESPHomeStreamConfig) {
	var port string

	streamCmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration, err := resolveESPHomeConfiguration(args, 0)
			if err != nil {
				return err
			}
			textMode := getTextMode()

			esClient, err := getESPHomeClient()
			if err != nil {
				return err
			}

			return streamToOutput(esClient, cfg.Operation, configuration, port, textMode)
		},
	}

	if cfg.HasPort {
		streamCmd.Flags().StringVar(&port, "port", "OTA", "Connection port: OTA (network) or serial port path")
	}

	esphomeCmd.AddCommand(streamCmd)
}

func init() {
	RegisterESPHomeStream(ESPHomeStreamConfig{
		Use:   "build <configuration>",
		Short: "Compile ESPHome firmware",
		Long: `Compile firmware for an ESPHome device configuration.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Output is streamed in real-time as the build progresses.`,
		Operation: "build",
	})

	RegisterESPHomeStream(ESPHomeStreamConfig{
		Use:   "upload <configuration>",
		Short: "Upload firmware to ESPHome device",
		Long: `Flash compiled firmware to an ESPHome device.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Use --port to specify the connection method: "OTA" for network (default),
or a serial port path like "/dev/ttyUSB0".

The firmware must have been compiled first using 'esphome build'.
Output is streamed in real-time.`,
		Operation: "upload",
		HasPort:   true,
	})

	RegisterESPHomeStream(ESPHomeStreamConfig{
		Use:   "run <configuration>",
		Short: "Build and upload firmware to ESPHome device",
		Long: `Compile and flash firmware to an ESPHome device in one step.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Use --port to specify the connection method: "OTA" for network (default),
or a serial port path like "/dev/ttyUSB0".

This is equivalent to running 'esphome build' followed by 'esphome upload'.
Output is streamed in real-time.`,
		Operation: "run",
		HasPort:   true,
	})

	RegisterESPHomeStream(ESPHomeStreamConfig{
		Use:   "logs <configuration>",
		Short: "Stream live device logs",
		Long: `Stream live logs from an ESPHome device.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Use --port to specify the connection method: "OTA" for network (default),
or a serial port path like "/dev/ttyUSB0".

Logs are streamed in real-time until interrupted (Ctrl+C).`,
		Operation: "logs",
		HasPort:   true,
	})
}

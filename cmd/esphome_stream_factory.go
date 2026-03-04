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
	Use     string // cobra command Use field, e.g. "build <configuration>"
	Short   string // one-line description
	Long    string // detailed description
	WSPath  string // WebSocket endpoint path, e.g. "/compile"
	HasPort bool   // whether to add --port flag
	// ExtraFields are additional key-value pairs merged into the spawn message.
	ExtraFields map[string]interface{}
}

// RegisterESPHomeStream generates and registers a streaming subcommand on
// esphomeCmd.
func RegisterESPHomeStream(cfg ESPHomeStreamConfig) {
	var port string

	streamCmd := &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if cfg.HasPort {
				spawnMsg["port"] = port
			}

			for k, v := range cfg.ExtraFields {
				spawnMsg[k] = v
			}

			return streamToOutput(esClient, cfg.WSPath, spawnMsg, textMode)
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
		WSPath:      "/compile",
		ExtraFields: map[string]interface{}{"only_generate": false},
	})

	RegisterESPHomeStream(ESPHomeStreamConfig{
		Use:   "validate <configuration>",
		Short: "Validate ESPHome device configuration",
		Long: `Validate an ESPHome device configuration without compiling.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Output is streamed in real-time.`,
		WSPath: "/validate",
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
		WSPath:  "/upload",
		HasPort: true,
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
		WSPath:  "/run",
		HasPort: true,
	})

	RegisterESPHomeStream(ESPHomeStreamConfig{
		Use:   "logs <configuration>",
		Short: "Stream live device logs",
		Long: `Stream live logs from an ESPHome device.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Use --port to specify the connection method: "OTA" for network (default),
or a serial port path like "/dev/ttyUSB0".

Logs are streamed in real-time until interrupted (Ctrl+C).`,
		WSPath:  "/logs",
		HasPort: true,
	})
}

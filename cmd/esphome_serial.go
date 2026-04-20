package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/home-assistant/hab/internal/esphomerecovery"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	esphomeSerialPort    string
	esphomeSerialChip    string
	esphomeSerialTool    string
	esphomeSerialTimeout time.Duration
	esphomeSerialForce   bool
)

var esphomeSerialCmd = &cobra.Command{
	Use:   "serial",
	Short: "Serial diagnostics and recovery helpers",
}

var esphomeSerialPortsCmd = &cobra.Command{
	Use:   "ports",
	Short: "List available serial ports from ESPHome dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()

		esClient, err := getESPHomeClient()
		if err != nil {
			return err
		}

		ports, err := esClient.GetSerialPorts()
		if err != nil {
			return err
		}

		output.PrintOutput(ports, textMode, "")
		return nil
	},
}

var esphomeSerialProbeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Run esptool flash probe on a serial port",
	RunE: func(cmd *cobra.Command, args []string) error {
		if esphomeSerialPort == "" {
			return fmt.Errorf("--port is required")
		}
		textMode := getTextMode()

		ctx, cancel := context.WithTimeout(context.Background(), esphomeSerialTimeout)
		defer cancel()

		result, err := esphomerecovery.Probe(ctx, esphomerecovery.Options{
			Port: esphomeSerialPort,
			Chip: esphomeSerialChip,
			Tool: esphomeSerialTool,
		})
		if err != nil {
			return err
		}

		output.PrintOutput(map[string]any{
			"operation": "probe",
			"port":      esphomeSerialPort,
			"chip":      esphomeSerialChip,
			"result":    result,
		}, textMode, "")
		return nil
	},
}

var esphomeSerialEraseCmd = &cobra.Command{
	Use:   "erase-flash",
	Short: "Erase flash via esptool (destructive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if esphomeSerialPort == "" {
			return fmt.Errorf("--port is required")
		}
		textMode := getTextMode()
		if err := confirmAction(esphomeSerialForce, fmt.Sprintf("Erase flash on %s?", esphomeSerialPort), "erase ESPHome device flash"); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), esphomeSerialTimeout)
		defer cancel()

		result, err := esphomerecovery.EraseFlash(ctx, esphomerecovery.Options{
			Port: esphomeSerialPort,
			Chip: esphomeSerialChip,
			Tool: esphomeSerialTool,
		})
		if err != nil {
			return err
		}

		output.PrintOutput(map[string]any{
			"operation": "erase_flash",
			"port":      esphomeSerialPort,
			"chip":      esphomeSerialChip,
			"result":    result,
		}, textMode, "")
		return nil
	},
}

func init() {
	esphomeCmd.AddCommand(esphomeSerialCmd)
	esphomeSerialCmd.AddCommand(esphomeSerialPortsCmd)
	esphomeSerialCmd.AddCommand(esphomeSerialProbeCmd)
	esphomeSerialCmd.AddCommand(esphomeSerialEraseCmd)

	for _, serialCmd := range []*cobra.Command{esphomeSerialProbeCmd, esphomeSerialEraseCmd} {
		serialCmd.Flags().StringVar(&esphomeSerialPort, "port", "", "Serial port path (e.g. /dev/ttyUSB0 or COM3)")
		serialCmd.Flags().StringVar(&esphomeSerialChip, "chip", "auto", "Target chip (auto, esp8266, esp32, esp32s2, esp32s3, esp32c3, esp32c6, esp32h2)")
		serialCmd.Flags().StringVar(&esphomeSerialTool, "tool", "", "Path to esptool binary or uvx (defaults to HAB_ESPTOOL_BIN, esptool, uvx esptool, or python -m esptool)")
		serialCmd.Flags().DurationVar(&esphomeSerialTimeout, "timeout", 2*time.Minute, "Command timeout")
	}

	esphomeSerialEraseCmd.Flags().BoolVar(&esphomeSerialForce, "force", false, "Skip erase confirmation")
}

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var esphomeConfigWriteFile string
var esphomeConfigWriteData string

var esphomeConfigWriteCmd = &cobra.Command{
	Use:   "config-write <configuration>",
	Short: "Write ESPHome device YAML configuration",
	Long: `Write or update the YAML configuration file for an ESPHome device.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Provide the YAML content via --data (inline string) or --file (path to local file).`,
	Args: cobra.ExactArgs(1),
	RunE: runESPHomeConfigWrite,
}

func init() {
	esphomeCmd.AddCommand(esphomeConfigWriteCmd)
	esphomeConfigWriteCmd.Flags().StringVarP(&esphomeConfigWriteData, "data", "d", "", "YAML content to write (inline)")
	esphomeConfigWriteCmd.Flags().StringVarP(&esphomeConfigWriteFile, "file", "f", "", "Path to YAML file to upload")
}

func runESPHomeConfigWrite(cmd *cobra.Command, args []string) error {
	configuration := args[0]
	textMode := getTextMode()

	var content string
	switch {
	case esphomeConfigWriteData == "-":
		// Read from stdin when --data - is specified
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		content = string(data)
	case esphomeConfigWriteData != "":
		content = esphomeConfigWriteData
	case esphomeConfigWriteFile != "":
		data, err := os.ReadFile(esphomeConfigWriteFile)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", esphomeConfigWriteFile, err)
		}
		content = string(data)
	default:
		return fmt.Errorf("provide YAML content via --data or --file")
	}

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	if err := esClient.WriteConfig(configuration, content); err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Configuration %s updated successfully.", configuration))
	return nil
}

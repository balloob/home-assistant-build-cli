package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var esphomeConfigReadCmd = &cobra.Command{
	Use:   "config-read <configuration>",
	Short: "Read ESPHome device YAML configuration",
	Long: `Read the YAML configuration file for an ESPHome device.

The configuration argument is the YAML filename (e.g. "living-room.yaml").
Returns the raw YAML content of the device configuration.`,
	Args: cobra.ExactArgs(1),
	RunE: runESPHomeConfigRead,
}

func init() {
	esphomeCmd.AddCommand(esphomeConfigReadCmd)
}

func runESPHomeConfigRead(cmd *cobra.Command, args []string) error {
	configuration := args[0]
	textMode := getTextMode()

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	content, err := esClient.ReadConfig(configuration)
	if err != nil {
		return err
	}

	if textMode {
		fmt.Print(content)
		return nil
	}

	output.PrintOutput(map[string]interface{}{
		"configuration": configuration,
		"content":       content,
	}, false, "")
	return nil
}

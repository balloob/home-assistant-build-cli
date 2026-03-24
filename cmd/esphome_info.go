package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var esphomeInfoCmd = &cobra.Command{
	Use:   "info [configuration]",
	Short: "Show ESPHome dashboard metadata for a configuration",
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

		info, err := esClient.GetInfo(configuration)
		if err != nil {
			return err
		}

		output.PrintOutput(info, textMode, "")
		return nil
	},
}

func init() {
	esphomeCmd.AddCommand(esphomeInfoCmd)
}

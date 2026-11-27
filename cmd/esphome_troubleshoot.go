package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var esphomeTroubleshootCmd = &cobra.Command{
	Use:   "troubleshoot [configuration]",
	Short: "Check why an ESPHome device is offline (DNS, mDNS, ping)",
	Long: `Run the ESPHome Device Builder's connectivity probe for a device: a live
DNS lookup of its address, an mDNS re-query, and a ping of the best address found.

Requires the ESPHome Device Builder; the classic dashboard has no equivalent.`,
	Args: cobra.MaximumNArgs(1),
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

		result, err := esClient.Troubleshoot(configuration)
		if err != nil {
			return err
		}

		output.PrintOutput(result, textMode, "")
		return nil
	},
}

func init() {
	esphomeCmd.AddCommand(esphomeTroubleshootCmd)
}

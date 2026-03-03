package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var systemInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get system information",
	Long:  `Get Home Assistant system information including version, location, and configuration.`,
	RunE:  runSystemInfo,
}

func init() {
	systemCmd.AddCommand(systemInfoCmd)
}

func runSystemInfo(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	config, err := restClient.GetConfig()
	if err != nil {
		return err
	}

	result := map[string]interface{}{
		"location_name": config["location_name"],
		"version":       config["version"],
		"state":         config["state"],
		"external_url":  config["external_url"],
		"internal_url":  config["internal_url"],
		"time_zone":     config["time_zone"],
		"unit_system":   config["unit_system"],
		"elevation":     config["elevation"],
		"latitude":      config["latitude"],
		"longitude":     config["longitude"],
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

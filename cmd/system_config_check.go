package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var systemConfigCheckCmd = &cobra.Command{
	Use:   "config-check",
	Short: "Validate configuration",
	Long:  `Validate the Home Assistant configuration files.`,
	RunE:  runSystemConfigCheck,
}

func init() {
	systemCmd.AddCommand(systemConfigCheckCmd)
}

func runSystemConfigCheck(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	result, err := restClient.CheckConfig()
	if err != nil {
		return err
	}

	if result["result"] == "valid" {
		output := map[string]interface{}{
			"valid":  true,
			"errors": nil,
		}
		client.PrintSuccess(output, textMode, "Configuration is valid.")
	} else {
		output := map[string]interface{}{
			"valid":  false,
			"errors": result["errors"],
		}
		client.PrintOutput(output, textMode, "")
	}

	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/output"
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
		data := map[string]interface{}{
			"valid":  true,
			"errors": nil,
		}
		output.PrintSuccess(data, textMode, "Configuration is valid.")
	} else {
		data := map[string]interface{}{
			"valid":  false,
			"errors": result["errors"],
		}
		output.PrintOutput(data, textMode, "")
	}

	return nil
}

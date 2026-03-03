package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
)

var (
	dashboardSaveConfigData   string
	dashboardSaveConfigFile   string
	dashboardSaveConfigFormat string
)

var dashboardSaveConfigCmd = &cobra.Command{
	Use:     "save-config <url_path>",
	Short:   "Save dashboard configuration",
	Long:    `Save the Lovelace configuration for a dashboard.`,
	GroupID: dashboardGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runDashboardSaveConfig,
}

func init() {
	dashboardCmd.AddCommand(dashboardSaveConfigCmd)
	dashboardSaveConfigCmd.Flags().StringVarP(&dashboardSaveConfigData, "data", "d", "", "Dashboard configuration as JSON")
	dashboardSaveConfigCmd.Flags().StringVarP(&dashboardSaveConfigFile, "file", "f", "", "Path to config file")
	dashboardSaveConfigCmd.Flags().StringVar(&dashboardSaveConfigFormat, "format", "", "Input format (json, yaml)")
}

func runDashboardSaveConfig(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	textMode := getTextMode()

	config, err := input.ParseInput(dashboardSaveConfigData, dashboardSaveConfigFile, dashboardSaveConfigFormat)
	if err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{
		"config": config,
	}
	// Use null for default lovelace dashboard
	if urlPath != "lovelace" {
		params["url_path"] = urlPath
	}

	_, err = ws.SendCommand("lovelace/config/save", params)
	if err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Dashboard config for '%s' saved.", urlPath))
	return nil
}

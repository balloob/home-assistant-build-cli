package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var dashboardGetCmd = &cobra.Command{
	Use:     "get <url_path>",
	Short:   "Get dashboard configuration",
	Long:    `Get the full configuration of a dashboard.`,
	GroupID: dashboardGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runDashboardGet,
}

func init() {
	dashboardCmd.AddCommand(dashboardGetCmd)
}

func runDashboardGet(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{}
	if urlPath != "lovelace" {
		params["url_path"] = urlPath
	}

	result, err := ws.SendCommand("lovelace/config", params)
	if err != nil {
		return err
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

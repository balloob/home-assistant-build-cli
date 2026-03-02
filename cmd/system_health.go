package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var systemHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Get system health status",
	Long:  `Get the health status of Home Assistant and its integrations.`,
	RunE:  runSystemHealth,
}

func init() {
	systemCmd.AddCommand(systemHealthCmd)
}

func runSystemHealth(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SystemHealthInfo()
	if err != nil {
		return err
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

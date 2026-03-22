package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var networkGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get network configuration",
	Long:  `Get configured network adapters and interface details.`,
	RunE:  runNetworkGet,
}

func init() {
	networkCmd.AddCommand(networkGetCmd)
}

func runNetworkGet(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.NetworkGet()
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

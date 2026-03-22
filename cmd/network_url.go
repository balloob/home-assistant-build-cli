package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var networkURLCmd = &cobra.Command{
	Use:   "url",
	Short: "Get network URLs",
	Long:  `Get internal, external, and cloud URLs for this Home Assistant instance.`,
	RunE:  runNetworkURL,
}

func init() {
	networkCmd.AddCommand(networkURLCmd)
}

func runNetworkURL(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.NetworkURL()
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show Marketplace status",
	Long: `Show the Marketplace status: categories, stage, GitHub connection,
and whether the current user accepted the warning.`,
	RunE: runMarketplaceInfo,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceInfoCmd)
}

func runMarketplaceInfo(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("marketplace/info", nil)
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

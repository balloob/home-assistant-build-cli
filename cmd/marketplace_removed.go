package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceRemovedCmd = &cobra.Command{
	Use:   "removed",
	Short: "List repositories removed from the Marketplace",
	Long: `List repositories that the Marketplace catalog removed, with the reason.
Ignored repositories are left out.`,
	RunE: runMarketplaceRemoved,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceRemovedCmd)
}

func runMarketplaceRemoved(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("marketplace/repositories/removed", nil)
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceGetCmd = &cobra.Command{
	Use:   "get <repository>",
	Short: "Get a Marketplace repository",
	Long: `Get the details of a Marketplace repository, including its README,
releases, installed version, and status.`,
	Example: `  hab marketplace get piitaya/lovelace-mushroom
  hab marketplace get 444350375`,
	Args: cobra.ExactArgs(1),
	RunE: runMarketplaceGet,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceGetCmd)
}

func runMarketplaceGet(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	repositoryID, err := resolveMarketplaceRepository(ws, args[0])
	if err != nil {
		return err
	}

	result, err := ws.SendCommand("marketplace/repository/info", map[string]interface{}{
		"repository_id": repositoryID,
	})
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceReleasesCmd = &cobra.Command{
	Use:   "releases <repository>",
	Short: "List the releases of a Marketplace repository",
	Long:  `List the GitHub releases of a Marketplace repository.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMarketplaceReleases,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceReleasesCmd)
}

func runMarketplaceReleases(cmd *cobra.Command, args []string) error {
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

	result, err := ws.SendCommand("marketplace/repository/releases", map[string]interface{}{
		"repository_id": repositoryID,
	})
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceClearNewCategories []string

var marketplaceClearNewCmd = &cobra.Command{
	Use:   "clear-new [repository]",
	Short: "Clear the new flag of Marketplace repositories",
	Long:  `Mark one repository, or all repositories in the given categories, as seen.`,
	Example: `  hab marketplace clear-new piitaya/lovelace-mushroom
  hab marketplace clear-new --category plugin --category theme`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMarketplaceClearNew,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceClearNewCmd)
	marketplaceClearNewCmd.Flags().StringSliceVar(&marketplaceClearNewCategories, "category", nil, "Category to clear (repeatable)")
	mergeSchemaAnnotation(marketplaceClearNewCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args", "flags"},
	})
}

func runMarketplaceClearNew(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	if len(args) == 0 && len(marketplaceClearNewCategories) == 0 {
		return fmt.Errorf("give a repository or at least one --category")
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{}
	if len(args) > 0 {
		repositoryID, err := resolveMarketplaceRepository(ws, args[0])
		if err != nil {
			return err
		}
		params["repository"] = repositoryID
	} else {
		params["categories"] = marketplaceClearNewCategories
	}

	if _, err := ws.SendCommand("marketplace/repositories/clear_new", params); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, "New flag cleared.", output.EnvelopeContext{Operation: "clear-new", ResourceType: "marketplace_repository"})
	return nil
}

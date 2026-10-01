package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceSetBetaCmd = &cobra.Command{
	Use:   "set-beta <repository> <true|false>",
	Short: "Show or hide beta versions of a Marketplace repository",
	Long:  `Choose whether the update entity of a repository offers pre-releases.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runMarketplaceSetBeta,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceSetBetaCmd)
	mergeSchemaAnnotation(marketplaceSetBetaCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "marketplace_repository",
		InputSources: []string{"args"},
	})
}

func runMarketplaceSetBeta(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	showBeta, err := strconv.ParseBool(args[1])
	if err != nil {
		return fmt.Errorf("invalid value %q: use true or false", args[1])
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	repositoryID, err := resolveMarketplaceRepository(ws, args[0])
	if err != nil {
		return err
	}

	if _, err := ws.SendCommand("marketplace/repository/beta", map[string]interface{}{
		"repository": repositoryID,
		"show_beta":  showBeta,
	}); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Beta versions of '%s' set to %t.", args[0], showBeta), output.EnvelopeContext{Operation: "set-beta", ResourceType: "marketplace_repository"})
	return nil
}

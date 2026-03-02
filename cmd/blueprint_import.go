package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var blueprintImportCmd = &cobra.Command{
	Use:   "import <url>",
	Short: "Import a blueprint from URL",
	Long:  `Import a blueprint from a URL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBlueprintImport,
}

func init() {
	blueprintCmd.AddCommand(blueprintImportCmd)
}

func runBlueprintImport(cmd *cobra.Command, args []string) error {
	url := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("blueprint/import", map[string]interface{}{
		"url": url,
	})
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Blueprint imported from %s", url))
	return nil
}

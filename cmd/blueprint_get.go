package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var blueprintGetPath string

var blueprintGetCmd = &cobra.Command{
	Use:   "get [path]",
	Short: "Get blueprint details",
	Long:  `Get details and inputs for a blueprint by its path. Use --domain to specify the domain (default: automation).`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBlueprintGet,
}

func init() {
	blueprintCmd.AddCommand(blueprintGetCmd)
	blueprintGetCmd.Flags().StringVar(&blueprintGetPath, "path", "", "Blueprint path to get")
	blueprintGetCmd.Flags().String("domain", "automation", "Domain of the blueprint (automation/script)")
}

func runBlueprintGet(cmd *cobra.Command, args []string) error {
	path, err := resolveArg(blueprintGetPath, args, 0, "blueprint path")
	if err != nil {
		return err
	}
	textMode := getTextMode()
	domain, _ := cmd.Flags().GetString("domain")

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	// First get the list to find the blueprint and its metadata
	listResult, err := ws.SendCommand("blueprint/list", map[string]interface{}{
		"domain": domain,
	})
	if err != nil {
		return err
	}

	// Extract the specific blueprint from the list
	if blueprints, ok := listResult.(map[string]interface{}); ok {
		if blueprint, ok := blueprints[path]; ok {
			result := map[string]interface{}{
				"path":   path,
				"domain": domain,
				"blueprint": blueprint,
			}
			output.PrintOutput(result, textMode, "")
			return nil
		}
	}

	// If not found in list format, return the path lookup result directly
	output.PrintOutput(map[string]interface{}{
		"path":   path,
		"domain": domain,
		"error":  "Blueprint not found",
	}, textMode, "")
	return nil
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	path := blueprintGetPath
	if path == "" && len(args) > 0 {
		path = args[0]
	}
	if path == "" {
		return fmt.Errorf("blueprint path is required (use --path flag or positional argument)")
	}
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")
	domain, _ := cmd.Flags().GetString("domain")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
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
			client.PrintOutput(result, textMode, "")
			return nil
		}
	}

	// If not found in list format, return the path lookup result directly
	client.PrintOutput(map[string]interface{}{
		"path":   path,
		"domain": domain,
		"error":  "Blueprint not found",
	}, textMode, "")
	return nil
}

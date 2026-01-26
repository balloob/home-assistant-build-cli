package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scriptGetCmd = &cobra.Command{
	Use:     "get <script_id>",
	Short:   "Get script configuration",
	Long:    `Get the full configuration of a script.`,
	GroupID: scriptGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runScriptGet,
}

func init() {
	scriptCmd.AddCommand(scriptGetCmd)
}

func runScriptGet(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	// Strip "script." prefix if provided - API expects just the ID
	scriptID = strings.TrimPrefix(scriptID, "script.")

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	result, err := restClient.Get("config/script/config/" + scriptID)
	if err != nil {
		return err
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

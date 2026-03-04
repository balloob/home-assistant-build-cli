package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var scriptUpdateInput InputFlags

var scriptUpdateCmd = &cobra.Command{
	Use:     "update <script_id>",
	Short:   "Update an existing script",
	Long:    `Update a script with new configuration.`,
	GroupID: scriptGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runScriptUpdate,
}

func init() {
	scriptCmd.AddCommand(scriptUpdateCmd)
	scriptUpdateInput.Register(scriptUpdateCmd)
}

func runScriptUpdate(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	// Strip "script." prefix if provided - API expects just the ID
	scriptID = strings.TrimPrefix(scriptID, "script.")

	textMode := getTextMode()

	config, err := scriptUpdateInput.Parse()
	if err != nil {
		return err
	}

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	result, err := restClient.Post("config/script/config/"+scriptID, config)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, "Script updated successfully.")
	return nil
}

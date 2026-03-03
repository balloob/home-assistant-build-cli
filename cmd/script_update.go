package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
)

var (
	scriptUpdateData   string
	scriptUpdateFile   string
	scriptUpdateFormat string
)

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
	scriptUpdateCmd.Flags().StringVarP(&scriptUpdateData, "data", "d", "", "Updated configuration as JSON")
	scriptUpdateCmd.Flags().StringVarP(&scriptUpdateFile, "file", "f", "", "Path to config file")
	scriptUpdateCmd.Flags().StringVar(&scriptUpdateFormat, "format", "", "Input format (json, yaml)")
}

func runScriptUpdate(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	// Strip "script." prefix if provided - API expects just the ID
	scriptID = strings.TrimPrefix(scriptID, "script.")

	textMode := getTextMode()

	config, err := input.ParseInput(scriptUpdateData, scriptUpdateFile, scriptUpdateFormat)
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

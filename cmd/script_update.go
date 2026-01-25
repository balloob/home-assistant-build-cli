package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scriptUpdateData   string
	scriptUpdateFile   string
	scriptUpdateFormat string
)

var scriptUpdateCmd = &cobra.Command{
	Use:   "update <script_id>",
	Short: "Update an existing script",
	Long:  `Update a script with new configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runScriptUpdate,
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

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	config, err := input.ParseInput(scriptUpdateData, scriptUpdateFile, scriptUpdateFormat)
	if err != nil {
		return err
	}

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	result, err := restClient.Post("config/script/config/"+scriptID, config)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, "Script updated successfully.")
	return nil
}

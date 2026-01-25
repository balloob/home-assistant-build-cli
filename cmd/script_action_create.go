package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scriptActionCreateData   string
	scriptActionCreateFile   string
	scriptActionCreateFormat string
)

var scriptActionCreateCmd = &cobra.Command{
	Use:   "create <script_id>",
	Short: "Create a new action",
	Long:  `Create a new action in a script's sequence.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runScriptActionCreate,
}

func init() {
	scriptActionCmd.AddCommand(scriptActionCreateCmd)
	scriptActionCreateCmd.Flags().StringVarP(&scriptActionCreateData, "data", "d", "", "Action configuration as JSON")
	scriptActionCreateCmd.Flags().StringVarP(&scriptActionCreateFile, "file", "f", "", "Path to config file")
	scriptActionCreateCmd.Flags().StringVar(&scriptActionCreateFormat, "format", "", "Input format (json, yaml)")
}

func runScriptActionCreate(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	scriptID = strings.TrimPrefix(scriptID, "script.")

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	actionConfig, err := input.ParseInput(scriptActionCreateData, scriptActionCreateFile, scriptActionCreateFormat)
	if err != nil {
		return err
	}

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	// Get current script config
	result, err := restClient.Get("config/script/config/" + scriptID)
	if err != nil {
		return err
	}

	config, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid script config")
	}

	// Get existing sequence
	sequence, ok := config["sequence"].([]interface{})
	if !ok {
		sequence = []interface{}{}
	}

	// Add the new action
	sequence = append(sequence, actionConfig)
	config["sequence"] = sequence

	// Save the config
	_, err = restClient.Post("config/script/config/"+scriptID, config)
	if err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  len(sequence) - 1,
		"config": actionConfig,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Action created at index %d.", len(sequence)-1))
	return nil
}

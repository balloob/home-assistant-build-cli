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
	automationActionCreateData   string
	automationActionCreateFile   string
	automationActionCreateFormat string
)

var automationActionCreateCmd = &cobra.Command{
	Use:   "create <automation_id>",
	Short: "Create a new action",
	Long:  `Create a new action in an automation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAutomationActionCreate,
}

func init() {
	automationActionCmd.AddCommand(automationActionCreateCmd)
	automationActionCreateCmd.Flags().StringVarP(&automationActionCreateData, "data", "d", "", "Action configuration as JSON")
	automationActionCreateCmd.Flags().StringVarP(&automationActionCreateFile, "file", "f", "", "Path to config file")
	automationActionCreateCmd.Flags().StringVar(&automationActionCreateFormat, "format", "", "Input format (json, yaml)")
}

func runAutomationActionCreate(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	automationID = strings.TrimPrefix(automationID, "automation.")

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	actionConfig, err := input.ParseInput(automationActionCreateData, automationActionCreateFile, automationActionCreateFormat)
	if err != nil {
		return err
	}

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	// Get current automation config
	result, err := restClient.Get("config/automation/config/" + automationID)
	if err != nil {
		return err
	}

	config, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid automation config")
	}

	// Get existing actions (try both keys)
	var actions []interface{}
	var actionKey string
	if a, ok := config["actions"].([]interface{}); ok {
		actions = a
		actionKey = "actions"
	} else if a, ok := config["action"].([]interface{}); ok {
		actions = a
		actionKey = "action"
	} else {
		actions = []interface{}{}
		actionKey = "actions"
	}

	// Add the new action
	actions = append(actions, actionConfig)
	config[actionKey] = actions

	// Save the config
	_, err = restClient.Post("config/automation/config/"+automationID, config)
	if err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  len(actions) - 1,
		"config": actionConfig,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Action created at index %d.", len(actions)-1))
	return nil
}

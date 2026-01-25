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
	automationTriggerCreateData   string
	automationTriggerCreateFile   string
	automationTriggerCreateFormat string
)

var automationTriggerCreateCmd = &cobra.Command{
	Use:   "create <automation_id>",
	Short: "Create a new trigger",
	Long:  `Create a new trigger in an automation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAutomationTriggerCreate,
}

func init() {
	automationTriggerParentCmd.AddCommand(automationTriggerCreateCmd)
	automationTriggerCreateCmd.Flags().StringVarP(&automationTriggerCreateData, "data", "d", "", "Trigger configuration as JSON")
	automationTriggerCreateCmd.Flags().StringVarP(&automationTriggerCreateFile, "file", "f", "", "Path to config file")
	automationTriggerCreateCmd.Flags().StringVar(&automationTriggerCreateFormat, "format", "", "Input format (json, yaml)")
}

func runAutomationTriggerCreate(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	automationID = strings.TrimPrefix(automationID, "automation.")

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	triggerConfig, err := input.ParseInput(automationTriggerCreateData, automationTriggerCreateFile, automationTriggerCreateFormat)
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

	// Get existing triggers (try both keys)
	var triggers []interface{}
	var triggerKey string
	if t, ok := config["triggers"].([]interface{}); ok {
		triggers = t
		triggerKey = "triggers"
	} else if t, ok := config["trigger"].([]interface{}); ok {
		triggers = t
		triggerKey = "trigger"
	} else {
		triggers = []interface{}{}
		triggerKey = "triggers"
	}

	// Add the new trigger
	triggers = append(triggers, triggerConfig)
	config[triggerKey] = triggers

	// Save the config
	_, err = restClient.Post("config/automation/config/"+automationID, config)
	if err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  len(triggers) - 1,
		"config": triggerConfig,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Trigger created at index %d.", len(triggers)-1))
	return nil
}

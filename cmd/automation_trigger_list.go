package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var automationTriggerListCmd = &cobra.Command{
	Use:   "list <automation_id>",
	Short: "List triggers in an automation",
	Long:  `List all triggers in an automation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAutomationTriggerList,
}

func init() {
	automationTriggerParentCmd.AddCommand(automationTriggerListCmd)
}

func runAutomationTriggerList(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	automationID = strings.TrimPrefix(automationID, "automation.")

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	result, err := restClient.Get("config/automation/config/" + automationID)
	if err != nil {
		return err
	}

	config, ok := result.(map[string]interface{})
	if !ok {
		client.PrintOutput([]interface{}{}, textMode, "")
		return nil
	}

	// Try both "triggers" and "trigger" keys (HA supports both)
	triggers, ok := config["triggers"].([]interface{})
	if !ok {
		triggers, ok = config["trigger"].([]interface{})
		if !ok {
			client.PrintOutput([]interface{}{}, textMode, "")
			return nil
		}
	}

	// Add index to each trigger for easier reference
	triggerList := make([]map[string]interface{}, len(triggers))
	for i, t := range triggers {
		triggerData := make(map[string]interface{})
		if trigger, ok := t.(map[string]interface{}); ok {
			for k, val := range trigger {
				triggerData[k] = val
			}
		}
		triggerData["index"] = i
		triggerList[i] = triggerData
	}

	client.PrintOutput(triggerList, textMode, "")
	return nil
}

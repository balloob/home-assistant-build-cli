package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var automationActionListCmd = &cobra.Command{
	Use:   "list <automation_id>",
	Short: "List actions in an automation",
	Long:  `List all actions in an automation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAutomationActionList,
}

func init() {
	automationActionCmd.AddCommand(automationActionListCmd)
}

func runAutomationActionList(cmd *cobra.Command, args []string) error {
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

	// Try both "actions" and "action" keys (HA supports both)
	actions, ok := config["actions"].([]interface{})
	if !ok {
		actions, ok = config["action"].([]interface{})
		if !ok {
			client.PrintOutput([]interface{}{}, textMode, "")
			return nil
		}
	}

	// Add index to each action for easier reference
	actionList := make([]map[string]interface{}, len(actions))
	for i, a := range actions {
		actionData := make(map[string]interface{})
		if action, ok := a.(map[string]interface{}); ok {
			for k, val := range action {
				actionData[k] = val
			}
		}
		actionData["index"] = i
		actionList[i] = actionData
	}

	client.PrintOutput(actionList, textMode, "")
	return nil
}

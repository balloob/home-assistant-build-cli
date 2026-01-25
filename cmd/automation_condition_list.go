package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var automationConditionListCmd = &cobra.Command{
	Use:   "list <automation_id>",
	Short: "List conditions in an automation",
	Long:  `List all conditions in an automation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAutomationConditionList,
}

func init() {
	automationConditionCmd.AddCommand(automationConditionListCmd)
}

func runAutomationConditionList(cmd *cobra.Command, args []string) error {
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

	// Try both "conditions" and "condition" keys (HA supports both)
	conditions, ok := config["conditions"].([]interface{})
	if !ok {
		conditions, ok = config["condition"].([]interface{})
		if !ok {
			client.PrintOutput([]interface{}{}, textMode, "")
			return nil
		}
	}

	// Add index to each condition for easier reference
	conditionList := make([]map[string]interface{}, len(conditions))
	for i, c := range conditions {
		conditionData := make(map[string]interface{})
		if condition, ok := c.(map[string]interface{}); ok {
			for k, val := range condition {
				conditionData[k] = val
			}
		}
		conditionData["index"] = i
		conditionList[i] = conditionData
	}

	client.PrintOutput(conditionList, textMode, "")
	return nil
}

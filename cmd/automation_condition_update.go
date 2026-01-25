package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	automationConditionUpdateData   string
	automationConditionUpdateFile   string
	automationConditionUpdateFormat string
)

var automationConditionUpdateCmd = &cobra.Command{
	Use:   "update <automation_id> <condition_index>",
	Short: "Update a condition",
	Long:  `Update a condition in an automation by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runAutomationConditionUpdate,
}

func init() {
	automationConditionCmd.AddCommand(automationConditionUpdateCmd)
	automationConditionUpdateCmd.Flags().StringVarP(&automationConditionUpdateData, "data", "d", "", "Condition configuration as JSON (replaces entire condition)")
	automationConditionUpdateCmd.Flags().StringVarP(&automationConditionUpdateFile, "file", "f", "", "Path to config file")
	automationConditionUpdateCmd.Flags().StringVar(&automationConditionUpdateFormat, "format", "", "Input format (json, yaml)")
}

func runAutomationConditionUpdate(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	automationID = strings.TrimPrefix(automationID, "automation.")
	conditionIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid condition index: %s", args[1])
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	newCondition, err := input.ParseInput(automationConditionUpdateData, automationConditionUpdateFile, automationConditionUpdateFormat)
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

	// Get existing conditions (try both keys)
	var conditions []interface{}
	var conditionKey string
	if c, ok := config["conditions"].([]interface{}); ok {
		conditions = c
		conditionKey = "conditions"
	} else if c, ok := config["condition"].([]interface{}); ok {
		conditions = c
		conditionKey = "condition"
	} else {
		return fmt.Errorf("no conditions in automation")
	}

	if conditionIndex < 0 || conditionIndex >= len(conditions) {
		return fmt.Errorf("condition index %d out of range (0-%d)", conditionIndex, len(conditions)-1)
	}

	// Update the condition
	conditions[conditionIndex] = newCondition
	config[conditionKey] = conditions

	// Save the config
	_, err = restClient.Post("config/automation/config/"+automationID, config)
	if err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  conditionIndex,
		"config": newCondition,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Condition at index %d updated.", conditionIndex))
	return nil
}

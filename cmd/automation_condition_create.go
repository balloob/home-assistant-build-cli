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
	automationConditionCreateData   string
	automationConditionCreateFile   string
	automationConditionCreateFormat string
)

var automationConditionCreateCmd = &cobra.Command{
	Use:   "create <automation_id>",
	Short: "Create a new condition",
	Long:  `Create a new condition in an automation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAutomationConditionCreate,
}

func init() {
	automationConditionCmd.AddCommand(automationConditionCreateCmd)
	automationConditionCreateCmd.Flags().StringVarP(&automationConditionCreateData, "data", "d", "", "Condition configuration as JSON")
	automationConditionCreateCmd.Flags().StringVarP(&automationConditionCreateFile, "file", "f", "", "Path to config file")
	automationConditionCreateCmd.Flags().StringVar(&automationConditionCreateFormat, "format", "", "Input format (json, yaml)")
}

func runAutomationConditionCreate(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	automationID = strings.TrimPrefix(automationID, "automation.")

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	conditionConfig, err := input.ParseInput(automationConditionCreateData, automationConditionCreateFile, automationConditionCreateFormat)
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
		conditions = []interface{}{}
		conditionKey = "conditions"
	}

	// Add the new condition
	conditions = append(conditions, conditionConfig)
	config[conditionKey] = conditions

	// Save the config
	_, err = restClient.Post("config/automation/config/"+automationID, config)
	if err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  len(conditions) - 1,
		"config": conditionConfig,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Condition created at index %d.", len(conditions)-1))
	return nil
}

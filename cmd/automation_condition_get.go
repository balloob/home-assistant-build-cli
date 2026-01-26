package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	automationConditionGetAutomationID string
	automationConditionGetIndex        int
)

var automationConditionGetCmd = &cobra.Command{
	Use:   "get [automation_id] [condition_index]",
	Short: "Get a specific condition",
	Long:  `Get a specific condition from an automation by index.`,
	Args:  cobra.MaximumNArgs(2),
	RunE:  runAutomationConditionGet,
}

func init() {
	automationConditionCmd.AddCommand(automationConditionGetCmd)
	automationConditionGetCmd.Flags().StringVar(&automationConditionGetAutomationID, "automation", "", "Automation ID")
	automationConditionGetCmd.Flags().IntVar(&automationConditionGetIndex, "index", -1, "Condition index")
}

func runAutomationConditionGet(cmd *cobra.Command, args []string) error {
	automationID := automationConditionGetAutomationID
	if automationID == "" && len(args) > 0 {
		automationID = args[0]
	}
	if automationID == "" {
		return fmt.Errorf("automation ID is required (use --automation flag or first positional argument)")
	}
	automationID = strings.TrimPrefix(automationID, "automation.")

	conditionIndex := automationConditionGetIndex
	if conditionIndex < 0 && len(args) > 1 {
		var err error
		conditionIndex, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid condition index: %s", args[1])
		}
	}
	if conditionIndex < 0 {
		return fmt.Errorf("condition index is required (use --index flag or second positional argument)")
	}

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
		return fmt.Errorf("invalid automation config")
	}

	// Try both "conditions" and "condition" keys
	conditions, ok := config["conditions"].([]interface{})
	if !ok {
		conditions, ok = config["condition"].([]interface{})
		if !ok {
			return fmt.Errorf("no conditions in automation")
		}
	}

	if conditionIndex < 0 || conditionIndex >= len(conditions) {
		return fmt.Errorf("condition index %d out of range (0-%d)", conditionIndex, len(conditions)-1)
	}

	condition := conditions[conditionIndex]
	conditionData := make(map[string]interface{})
	if c, ok := condition.(map[string]interface{}); ok {
		for k, val := range c {
			conditionData[k] = val
		}
	}
	conditionData["index"] = conditionIndex

	client.PrintOutput(conditionData, textMode, "")
	return nil
}

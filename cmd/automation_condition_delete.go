package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var automationConditionDeleteForce bool

var automationConditionDeleteCmd = &cobra.Command{
	Use:   "delete <automation_id> <condition_index>",
	Short: "Delete a condition",
	Long:  `Delete a condition from an automation by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runAutomationConditionDelete,
}

func init() {
	automationConditionCmd.AddCommand(automationConditionDeleteCmd)
	automationConditionDeleteCmd.Flags().BoolVarP(&automationConditionDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runAutomationConditionDelete(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	automationID = strings.TrimPrefix(automationID, "automation.")
	conditionIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid condition index: %s", args[1])
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

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

	// Confirmation prompt
	if !automationConditionDeleteForce && !textMode {
		fmt.Printf("Are you sure you want to delete condition at index %d? [y/N]: ", conditionIndex)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
	}

	// Remove the condition
	conditions = append(conditions[:conditionIndex], conditions[conditionIndex+1:]...)
	config[conditionKey] = conditions

	// Save the config
	_, err = restClient.Post("config/automation/config/"+automationID, config)
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Condition at index %d deleted.", conditionIndex))
	return nil
}

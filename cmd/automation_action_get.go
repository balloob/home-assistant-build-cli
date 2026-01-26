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
	automationActionGetAutomationID string
	automationActionGetIndex        int
)

var automationActionGetCmd = &cobra.Command{
	Use:   "get [automation_id] [action_index]",
	Short: "Get a specific action",
	Long:  `Get a specific action from an automation by index.`,
	Args:  cobra.MaximumNArgs(2),
	RunE:  runAutomationActionGet,
}

func init() {
	automationActionCmd.AddCommand(automationActionGetCmd)
	automationActionGetCmd.Flags().StringVar(&automationActionGetAutomationID, "automation", "", "Automation ID")
	automationActionGetCmd.Flags().IntVar(&automationActionGetIndex, "index", -1, "Action index")
}

func runAutomationActionGet(cmd *cobra.Command, args []string) error {
	automationID := automationActionGetAutomationID
	if automationID == "" && len(args) > 0 {
		automationID = args[0]
	}
	if automationID == "" {
		return fmt.Errorf("automation ID is required (use --automation flag or first positional argument)")
	}
	automationID = strings.TrimPrefix(automationID, "automation.")

	actionIndex := automationActionGetIndex
	if actionIndex < 0 && len(args) > 1 {
		var err error
		actionIndex, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid action index: %s", args[1])
		}
	}
	if actionIndex < 0 {
		return fmt.Errorf("action index is required (use --index flag or second positional argument)")
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

	// Try both "actions" and "action" keys
	actions, ok := config["actions"].([]interface{})
	if !ok {
		actions, ok = config["action"].([]interface{})
		if !ok {
			return fmt.Errorf("no actions in automation")
		}
	}

	if actionIndex < 0 || actionIndex >= len(actions) {
		return fmt.Errorf("action index %d out of range (0-%d)", actionIndex, len(actions)-1)
	}

	action := actions[actionIndex]
	actionData := make(map[string]interface{})
	if a, ok := action.(map[string]interface{}); ok {
		for k, val := range a {
			actionData[k] = val
		}
	}
	actionData["index"] = actionIndex

	client.PrintOutput(actionData, textMode, "")
	return nil
}

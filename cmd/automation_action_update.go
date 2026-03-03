package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
)

var (
	automationActionUpdateData   string
	automationActionUpdateFile   string
	automationActionUpdateFormat string
)

var automationActionUpdateCmd = &cobra.Command{
	Use:   "update <automation_id> <action_index>",
	Short: "Update an action",
	Long:  `Update an action in an automation by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runAutomationActionUpdate,
}

func init() {
	automationActionCmd.AddCommand(automationActionUpdateCmd)
	automationActionUpdateCmd.Flags().StringVarP(&automationActionUpdateData, "data", "d", "", "Action configuration as JSON (replaces entire action)")
	automationActionUpdateCmd.Flags().StringVarP(&automationActionUpdateFile, "file", "f", "", "Path to config file")
	automationActionUpdateCmd.Flags().StringVar(&automationActionUpdateFormat, "format", "", "Input format (json, yaml)")
}

func runAutomationActionUpdate(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	actionIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid action index: %s", args[1])
	}

	textMode := getTextMode()

	newAction, err := input.ParseInput(automationActionUpdateData, automationActionUpdateFile, automationActionUpdateFormat)
	if err != nil {
		return err
	}

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	configID, err := resolveAutomationConfigID(restClient, automationID)
	if err != nil {
		return err
	}

	// Get current automation config
	result, err := restClient.Get("config/automation/config/" + configID)
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
		return fmt.Errorf("no actions in automation")
	}

	if actionIndex < 0 || actionIndex >= len(actions) {
		return fmt.Errorf("action index %d out of range (0-%d)", actionIndex, len(actions)-1)
	}

	// Update the action
	actions[actionIndex] = newAction
	config[actionKey] = actions

	// Save the config
	_, err = restClient.Post("config/automation/config/"+configID, config)
	if err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  actionIndex,
		"config": newAction,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Action at index %d updated.", actionIndex))
	return nil
}

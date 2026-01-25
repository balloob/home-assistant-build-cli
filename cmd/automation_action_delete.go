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

var automationActionDeleteForce bool

var automationActionDeleteCmd = &cobra.Command{
	Use:   "delete <automation_id> <action_index>",
	Short: "Delete an action",
	Long:  `Delete an action from an automation by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runAutomationActionDelete,
}

func init() {
	automationActionCmd.AddCommand(automationActionDeleteCmd)
	automationActionDeleteCmd.Flags().BoolVarP(&automationActionDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runAutomationActionDelete(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	automationID = strings.TrimPrefix(automationID, "automation.")
	actionIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid action index: %s", args[1])
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

	// Confirmation prompt
	if !automationActionDeleteForce && !textMode {
		fmt.Printf("Are you sure you want to delete action at index %d? [y/N]: ", actionIndex)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
	}

	// Remove the action
	actions = append(actions[:actionIndex], actions[actionIndex+1:]...)
	config[actionKey] = actions

	// Save the config
	_, err = restClient.Post("config/automation/config/"+automationID, config)
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Action at index %d deleted.", actionIndex))
	return nil
}

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

var automationTriggerDeleteForce bool

var automationTriggerDeleteCmd = &cobra.Command{
	Use:   "delete <automation_id> <trigger_index>",
	Short: "Delete a trigger",
	Long:  `Delete a trigger from an automation by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runAutomationTriggerDelete,
}

func init() {
	automationTriggerParentCmd.AddCommand(automationTriggerDeleteCmd)
	automationTriggerDeleteCmd.Flags().BoolVarP(&automationTriggerDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runAutomationTriggerDelete(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	automationID = strings.TrimPrefix(automationID, "automation.")
	triggerIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid trigger index: %s", args[1])
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

	// Get existing triggers (try both keys)
	var triggers []interface{}
	var triggerKey string
	if t, ok := config["triggers"].([]interface{}); ok {
		triggers = t
		triggerKey = "triggers"
	} else if t, ok := config["trigger"].([]interface{}); ok {
		triggers = t
		triggerKey = "trigger"
	} else {
		return fmt.Errorf("no triggers in automation")
	}

	if triggerIndex < 0 || triggerIndex >= len(triggers) {
		return fmt.Errorf("trigger index %d out of range (0-%d)", triggerIndex, len(triggers)-1)
	}

	// Confirmation prompt
	if !automationTriggerDeleteForce && !textMode {
		fmt.Printf("Are you sure you want to delete trigger at index %d? [y/N]: ", triggerIndex)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
	}

	// Remove the trigger
	triggers = append(triggers[:triggerIndex], triggers[triggerIndex+1:]...)
	config[triggerKey] = triggers

	// Save the config
	_, err = restClient.Post("config/automation/config/"+automationID, config)
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Trigger at index %d deleted.", triggerIndex))
	return nil
}

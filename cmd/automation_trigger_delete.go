package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
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
	triggerIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid trigger index: %s", args[1])
	}

	textMode := getTextMode()

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
	_, err = restClient.Post("config/automation/config/"+configID, config)
	if err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Trigger at index %d deleted.", triggerIndex))
	return nil
}

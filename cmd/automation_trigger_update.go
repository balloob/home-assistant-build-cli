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
	automationTriggerUpdateData   string
	automationTriggerUpdateFile   string
	automationTriggerUpdateFormat string
)

var automationTriggerUpdateCmd = &cobra.Command{
	Use:   "update <automation_id> <trigger_index>",
	Short: "Update a trigger",
	Long:  `Update a trigger in an automation by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runAutomationTriggerUpdate,
}

func init() {
	automationTriggerParentCmd.AddCommand(automationTriggerUpdateCmd)
	automationTriggerUpdateCmd.Flags().StringVarP(&automationTriggerUpdateData, "data", "d", "", "Trigger configuration as JSON (replaces entire trigger)")
	automationTriggerUpdateCmd.Flags().StringVarP(&automationTriggerUpdateFile, "file", "f", "", "Path to config file")
	automationTriggerUpdateCmd.Flags().StringVar(&automationTriggerUpdateFormat, "format", "", "Input format (json, yaml)")
}

func runAutomationTriggerUpdate(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	automationID = strings.TrimPrefix(automationID, "automation.")
	triggerIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid trigger index: %s", args[1])
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	newTrigger, err := input.ParseInput(automationTriggerUpdateData, automationTriggerUpdateFile, automationTriggerUpdateFormat)
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

	// Update the trigger
	triggers[triggerIndex] = newTrigger
	config[triggerKey] = triggers

	// Save the config
	_, err = restClient.Post("config/automation/config/"+automationID, config)
	if err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  triggerIndex,
		"config": newTrigger,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Trigger at index %d updated.", triggerIndex))
	return nil
}

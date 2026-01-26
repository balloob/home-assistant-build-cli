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
	automationTriggerGetAutomationID string
	automationTriggerGetIndex        int
)

var automationTriggerGetCmd = &cobra.Command{
	Use:   "get [automation_id] [trigger_index]",
	Short: "Get a specific trigger",
	Long:  `Get a specific trigger from an automation by index.`,
	Args:  cobra.MaximumNArgs(2),
	RunE:  runAutomationTriggerGet,
}

func init() {
	automationTriggerParentCmd.AddCommand(automationTriggerGetCmd)
	automationTriggerGetCmd.Flags().StringVar(&automationTriggerGetAutomationID, "automation", "", "Automation ID")
	automationTriggerGetCmd.Flags().IntVar(&automationTriggerGetIndex, "index", -1, "Trigger index")
}

func runAutomationTriggerGet(cmd *cobra.Command, args []string) error {
	automationID := automationTriggerGetAutomationID
	if automationID == "" && len(args) > 0 {
		automationID = args[0]
	}
	if automationID == "" {
		return fmt.Errorf("automation ID is required (use --automation flag or first positional argument)")
	}
	automationID = strings.TrimPrefix(automationID, "automation.")

	triggerIndex := automationTriggerGetIndex
	if triggerIndex < 0 && len(args) > 1 {
		var err error
		triggerIndex, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid trigger index: %s", args[1])
		}
	}
	if triggerIndex < 0 {
		return fmt.Errorf("trigger index is required (use --index flag or second positional argument)")
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

	// Try both "triggers" and "trigger" keys
	triggers, ok := config["triggers"].([]interface{})
	if !ok {
		triggers, ok = config["trigger"].([]interface{})
		if !ok {
			return fmt.Errorf("no triggers in automation")
		}
	}

	if triggerIndex < 0 || triggerIndex >= len(triggers) {
		return fmt.Errorf("trigger index %d out of range (0-%d)", triggerIndex, len(triggers)-1)
	}

	trigger := triggers[triggerIndex]
	triggerData := make(map[string]interface{})
	if t, ok := trigger.(map[string]interface{}); ok {
		for k, val := range t {
			triggerData[k] = val
		}
	}
	triggerData["index"] = triggerIndex

	client.PrintOutput(triggerData, textMode, "")
	return nil
}

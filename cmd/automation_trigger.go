package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var automationTriggerSkipCondition bool

var automationTriggerCmd = &cobra.Command{
	Use:     "run <automation_id>",
	Short:   "Manually run an automation",
	Long:    `Manually run an automation (triggers it).`,
	GroupID: automationGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runAutomationTrigger,
}

func init() {
	automationCmd.AddCommand(automationTriggerCmd)
	automationTriggerCmd.Flags().BoolVar(&automationTriggerSkipCondition, "skip-condition", false, "Skip automation conditions")
}

func runAutomationTrigger(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	if !strings.HasPrefix(automationID, "automation.") {
		automationID = "automation." + automationID
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	serviceData := map[string]interface{}{
		"entity_id": automationID,
	}
	if automationTriggerSkipCondition {
		serviceData["skip_condition"] = true
	}

	_, err = restClient.CallService("automation", "trigger", serviceData)
	if err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Automation %s triggered.", automationID))
	return nil
}

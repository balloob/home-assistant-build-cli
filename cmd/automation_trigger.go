package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
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
	automationID := ensureDomainPrefix(args[0], "automation")

	textMode := getTextMode()

	restClient, err := getRESTClient()
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

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Automation %s triggered.", automationID))
	return nil
}

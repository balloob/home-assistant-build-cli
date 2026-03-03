package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var automationGetID string

var automationGetCmd = &cobra.Command{
	Use:     "get [automation_id]",
	Short:   "Get automation configuration",
	Long:    `Get the full configuration of an automation.`,
	GroupID: automationGroupCommands,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runAutomationGet,
}

func init() {
	automationCmd.AddCommand(automationGetCmd)
	automationGetCmd.Flags().StringVar(&automationGetID, "automation", "", "Automation ID to get")
}

func runAutomationGet(cmd *cobra.Command, args []string) error {
	automationID := automationGetID
	if automationID == "" && len(args) > 0 {
		automationID = args[0]
	}
	if automationID == "" {
		return fmt.Errorf("automation ID is required (use --automation flag or positional argument)")
	}
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	// Resolve entity slug → internal config ID (required in HA 2024.4+)
	configID, err := resolveAutomationConfigID(restClient, automationID)
	if err != nil {
		return err
	}

	result, err := restClient.Get("config/automation/config/" + configID)
	if err != nil {
		return err
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

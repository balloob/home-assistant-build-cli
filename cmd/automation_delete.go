package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var automationDeleteForce bool

var automationDeleteCmd = &cobra.Command{
	Use:     "delete <automation_id>",
	Short:   "Delete an automation",
	Long:    `Delete an automation from Home Assistant.`,
	GroupID: automationGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runAutomationDelete,
}

func init() {
	automationCmd.AddCommand(automationDeleteCmd)
	automationDeleteCmd.Flags().BoolVarP(&automationDeleteForce, "force", "f", false, "Skip confirmation")
}

func runAutomationDelete(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	configID, err := resolveAutomationConfigID(restClient, automationID)
	if err != nil {
		return err
	}

	if !confirmAction(automationDeleteForce, textMode, fmt.Sprintf("Delete automation %s?", automationID)) {
		fmt.Println("Cancelled.")
		return nil
	}

	_, err = restClient.Delete("config/automation/config/" + configID)
	if err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Automation %s deleted.", automationID))
	return nil
}

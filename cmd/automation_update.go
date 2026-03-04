package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var automationUpdateInput InputFlags

var automationUpdateCmd = &cobra.Command{
	Use:     "update <automation_id>",
	Short:   "Update an existing automation",
	Long:    `Update an automation with new configuration.`,
	GroupID: automationGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runAutomationUpdate,
}

func init() {
	automationCmd.AddCommand(automationUpdateCmd)
	automationUpdateInput.Register(automationUpdateCmd)
}

func runAutomationUpdate(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	textMode := getTextMode()

	config, err := automationUpdateInput.Parse()
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

	result, err := restClient.Post("config/automation/config/"+configID, config)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, "Automation updated successfully.")
	return nil
}

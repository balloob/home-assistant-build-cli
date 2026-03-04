package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var automationCreateInput InputFlags

var automationCreateCmd = &cobra.Command{
	Use:     "create <id>",
	Short:   "Create a new automation",
	Long:    `Create a new automation from JSON or YAML. The ID is used to identify the automation.`,
	GroupID: automationGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runAutomationCreate,
}

func init() {
	automationCmd.AddCommand(automationCreateCmd)
	automationCreateInput.Register(automationCreateCmd)
}

func runAutomationCreate(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	textMode := getTextMode()

	config, err := automationCreateInput.Parse()
	if err != nil {
		return err
	}

	if _, ok := config["alias"]; !ok {
		return fmt.Errorf("automation must have an 'alias' field")
	}

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	result, err := restClient.Post(fmt.Sprintf("config/automation/config/%s", automationID), config)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Automation %s created successfully.", automationID))
	return nil
}

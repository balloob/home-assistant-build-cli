package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
)

var (
	automationUpdateData   string
	automationUpdateFile   string
	automationUpdateFormat string
)

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
	automationUpdateCmd.Flags().StringVarP(&automationUpdateData, "data", "d", "", "Updated configuration as JSON")
	automationUpdateCmd.Flags().StringVarP(&automationUpdateFile, "file", "f", "", "Path to config file")
	automationUpdateCmd.Flags().StringVar(&automationUpdateFormat, "format", "", "Input format (json, yaml)")
}

func runAutomationUpdate(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	textMode := getTextMode()

	config, err := input.ParseInput(automationUpdateData, automationUpdateFile, automationUpdateFormat)
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

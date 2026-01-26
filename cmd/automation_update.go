package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	if !strings.HasPrefix(automationID, "automation.") {
		automationID = "automation." + automationID
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	config, err := input.ParseInput(automationUpdateData, automationUpdateFile, automationUpdateFormat)
	if err != nil {
		return err
	}

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	result, err := restClient.Post("config/automation/config/"+automationID, config)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, "Automation updated successfully.")
	return nil
}

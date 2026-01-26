package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var automationGetCmd = &cobra.Command{
	Use:     "get <automation_id>",
	Short:   "Get automation configuration",
	Long:    `Get the full configuration of an automation.`,
	GroupID: automationGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runAutomationGet,
}

func init() {
	automationCmd.AddCommand(automationGetCmd)
}

func runAutomationGet(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	// Strip "automation." prefix if provided - API expects just the ID
	automationID = strings.TrimPrefix(automationID, "automation.")

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

	client.PrintOutput(result, textMode, "")
	return nil
}

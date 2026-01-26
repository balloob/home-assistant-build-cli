package cmd

import (
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var automationTraceRunID string

var automationTraceCmd = &cobra.Command{
	Use:     "trace <automation_id>",
	Short:   "Get execution traces for debugging",
	Long:    `Get execution traces for an automation.`,
	GroupID: automationGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runAutomationTrace,
}

func init() {
	automationCmd.AddCommand(automationTraceCmd)
	automationTraceCmd.Flags().StringVar(&automationTraceRunID, "run-id", "", "Specific run ID to get trace for")
}

func runAutomationTrace(cmd *cobra.Command, args []string) error {
	automationID := args[0]
	if !strings.HasPrefix(automationID, "automation.") {
		automationID = "automation." + automationID
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	itemID := strings.TrimPrefix(automationID, "automation.")

	var result interface{}
	if automationTraceRunID != "" {
		result, err = ws.SendCommand("trace/get", map[string]interface{}{
			"domain":  "automation",
			"item_id": itemID,
			"run_id":  automationTraceRunID,
		})
	} else {
		result, err = ws.SendCommand("trace/list", map[string]interface{}{
			"domain":  "automation",
			"item_id": itemID,
		})
	}
	if err != nil {
		return err
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

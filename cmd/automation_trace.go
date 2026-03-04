package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	automationTraceRunID string
	automationTraceID    string
)

var automationTraceCmd = &cobra.Command{
	Use:     "trace [automation_id]",
	Short:   "Get execution traces for debugging",
	Long:    `Get execution traces for an automation.`,
	GroupID: automationGroupCommands,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runAutomationTrace,
}

func init() {
	automationCmd.AddCommand(automationTraceCmd)
	automationTraceCmd.Flags().StringVar(&automationTraceID, "automation", "", "Automation ID to get traces for")
	automationTraceCmd.Flags().StringVar(&automationTraceRunID, "run-id", "", "Specific run ID to get trace for")
}

func runAutomationTrace(cmd *cobra.Command, args []string) error {
	automationID, err := resolveArg(automationTraceID, args, 0, "automation ID")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(automationID, "automation.") {
		automationID = "automation." + automationID
	}

	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
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

	output.PrintOutput(result, textMode, "")
	return nil
}

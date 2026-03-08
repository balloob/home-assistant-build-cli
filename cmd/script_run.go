package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	scriptRunData string
	scriptRunID   string
)

var scriptRunCmd = &cobra.Command{
	Use:     "run [script_id]",
	Short:   "Execute a script",
	Long:    `Execute a script with optional variables.`,
	GroupID: scriptGroupCommands,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runScriptRun,
}

func init() {
	scriptCmd.AddCommand(scriptRunCmd)
	scriptRunCmd.Flags().StringVar(&scriptRunID, "script", "", "Script ID to execute")
	scriptRunCmd.Flags().StringVarP(&scriptRunData, "data", "d", "", "Script variables as JSON")
}

func runScriptRun(cmd *cobra.Command, args []string) error {
	scriptID, err := resolveArg(scriptRunID, args, 0, "script ID")
	if err != nil {
		return err
	}
	scriptID = ensureDomainPrefix(scriptID, "script")

	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	serviceData := make(map[string]interface{})
	serviceData["entity_id"] = scriptID

	if scriptRunData != "" {
		var variables map[string]interface{}
		if err := json.Unmarshal([]byte(scriptRunData), &variables); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}
		for k, v := range variables {
			serviceData[k] = v
		}
	}

	_, err = restClient.CallService("script", "turn_on", serviceData)
	if err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Script %s executed.", scriptID))
	return nil
}

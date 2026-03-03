package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var scriptActionListCmd = &cobra.Command{
	Use:   "list <script_id>",
	Short: "List actions in a script",
	Long:  `List all actions in a script's sequence.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runScriptActionList,
}

func init() {
	scriptActionCmd.AddCommand(scriptActionListCmd)
}

func runScriptActionList(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	scriptID = strings.TrimPrefix(scriptID, "script.")

	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	result, err := restClient.Get("config/script/config/" + scriptID)
	if err != nil {
		return err
	}

	config, ok := result.(map[string]interface{})
	if !ok {
		output.PrintOutput([]interface{}{}, textMode, "")
		return nil
	}

	// Scripts use "sequence" for actions
	sequence, ok := config["sequence"].([]interface{})
	if !ok {
		output.PrintOutput([]interface{}{}, textMode, "")
		return nil
	}

	// Add index to each action for easier reference
	actionList := make([]map[string]interface{}, len(sequence))
	for i, a := range sequence {
		actionData := make(map[string]interface{})
		if action, ok := a.(map[string]interface{}); ok {
			for k, val := range action {
				actionData[k] = val
			}
		}
		actionData["index"] = i
		actionList[i] = actionData
	}

	output.PrintOutput(actionList, textMode, "")
	return nil
}

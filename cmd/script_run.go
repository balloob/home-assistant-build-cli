package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	scriptID := scriptRunID
	if scriptID == "" && len(args) > 0 {
		scriptID = args[0]
	}
	if scriptID == "" {
		return fmt.Errorf("script ID is required (use --script flag or positional argument)")
	}
	if !strings.HasPrefix(scriptID, "script.") {
		scriptID = "script." + scriptID
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
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

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Script %s executed.", scriptID))
	return nil
}

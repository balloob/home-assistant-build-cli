package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scriptActionGetScriptID string
	scriptActionGetIndex    int
)

var scriptActionGetCmd = &cobra.Command{
	Use:   "get [script_id] [action_index]",
	Short: "Get a specific action",
	Long:  `Get a specific action from a script by index.`,
	Args:  cobra.MaximumNArgs(2),
	RunE:  runScriptActionGet,
}

func init() {
	scriptActionCmd.AddCommand(scriptActionGetCmd)
	scriptActionGetCmd.Flags().StringVar(&scriptActionGetScriptID, "script", "", "Script ID")
	scriptActionGetCmd.Flags().IntVar(&scriptActionGetIndex, "index", -1, "Action index")
}

func runScriptActionGet(cmd *cobra.Command, args []string) error {
	scriptID := scriptActionGetScriptID
	if scriptID == "" && len(args) > 0 {
		scriptID = args[0]
	}
	if scriptID == "" {
		return fmt.Errorf("script ID is required (use --script flag or first positional argument)")
	}
	scriptID = strings.TrimPrefix(scriptID, "script.")

	actionIndex := scriptActionGetIndex
	if actionIndex < 0 && len(args) > 1 {
		var err error
		actionIndex, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid action index: %s", args[1])
		}
	}
	if actionIndex < 0 {
		return fmt.Errorf("action index is required (use --index flag or second positional argument)")
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	result, err := restClient.Get("config/script/config/" + scriptID)
	if err != nil {
		return err
	}

	config, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid script config")
	}

	// Scripts use "sequence" for actions
	sequence, ok := config["sequence"].([]interface{})
	if !ok {
		return fmt.Errorf("no sequence in script")
	}

	if actionIndex < 0 || actionIndex >= len(sequence) {
		return fmt.Errorf("action index %d out of range (0-%d)", actionIndex, len(sequence)-1)
	}

	action := sequence[actionIndex]
	actionData := make(map[string]interface{})
	if a, ok := action.(map[string]interface{}); ok {
		for k, val := range a {
			actionData[k] = val
		}
	}
	actionData["index"] = actionIndex

	client.PrintOutput(actionData, textMode, "")
	return nil
}

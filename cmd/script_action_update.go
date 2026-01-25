package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scriptActionUpdateData   string
	scriptActionUpdateFile   string
	scriptActionUpdateFormat string
)

var scriptActionUpdateCmd = &cobra.Command{
	Use:   "update <script_id> <action_index>",
	Short: "Update an action",
	Long:  `Update an action in a script's sequence by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runScriptActionUpdate,
}

func init() {
	scriptActionCmd.AddCommand(scriptActionUpdateCmd)
	scriptActionUpdateCmd.Flags().StringVarP(&scriptActionUpdateData, "data", "d", "", "Action configuration as JSON (replaces entire action)")
	scriptActionUpdateCmd.Flags().StringVarP(&scriptActionUpdateFile, "file", "f", "", "Path to config file")
	scriptActionUpdateCmd.Flags().StringVar(&scriptActionUpdateFormat, "format", "", "Input format (json, yaml)")
}

func runScriptActionUpdate(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	scriptID = strings.TrimPrefix(scriptID, "script.")
	actionIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid action index: %s", args[1])
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	newAction, err := input.ParseInput(scriptActionUpdateData, scriptActionUpdateFile, scriptActionUpdateFormat)
	if err != nil {
		return err
	}

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	// Get current script config
	result, err := restClient.Get("config/script/config/" + scriptID)
	if err != nil {
		return err
	}

	config, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid script config")
	}

	// Get existing sequence
	sequence, ok := config["sequence"].([]interface{})
	if !ok {
		return fmt.Errorf("no sequence in script")
	}

	if actionIndex < 0 || actionIndex >= len(sequence) {
		return fmt.Errorf("action index %d out of range (0-%d)", actionIndex, len(sequence)-1)
	}

	// Update the action
	sequence[actionIndex] = newAction
	config["sequence"] = sequence

	// Save the config
	_, err = restClient.Post("config/script/config/"+scriptID, config)
	if err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  actionIndex,
		"config": newAction,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Action at index %d updated.", actionIndex))
	return nil
}

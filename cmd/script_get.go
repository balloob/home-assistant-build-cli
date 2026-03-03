package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var scriptGetID string

var scriptGetCmd = &cobra.Command{
	Use:     "get [script_id]",
	Short:   "Get script configuration",
	Long:    `Get the full configuration of a script.`,
	GroupID: scriptGroupCommands,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runScriptGet,
}

func init() {
	scriptCmd.AddCommand(scriptGetCmd)
	scriptGetCmd.Flags().StringVar(&scriptGetID, "script", "", "Script ID to get")
}

func runScriptGet(cmd *cobra.Command, args []string) error {
	scriptID := scriptGetID
	if scriptID == "" && len(args) > 0 {
		scriptID = args[0]
	}
	if scriptID == "" {
		return fmt.Errorf("script ID is required (use --script flag or positional argument)")
	}
	// Strip "script." prefix if provided - API expects just the ID
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

	output.PrintOutput(result, textMode, "")
	return nil
}

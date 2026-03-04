package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var scriptCreateInput InputFlags

var scriptCreateCmd = &cobra.Command{
	Use:     "create <id>",
	Short:   "Create a new script",
	Long:    `Create a new script from JSON or YAML. The ID is used to identify the script.`,
	GroupID: scriptGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runScriptCreate,
}

func init() {
	scriptCmd.AddCommand(scriptCreateCmd)
	scriptCreateInput.Register(scriptCreateCmd)
}

func runScriptCreate(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	textMode := getTextMode()

	config, err := scriptCreateInput.Parse()
	if err != nil {
		return err
	}

	if _, ok := config["alias"]; !ok {
		return fmt.Errorf("script must have an 'alias' field")
	}

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	result, err := restClient.Post(fmt.Sprintf("config/script/config/%s", scriptID), config)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Script %s created successfully.", scriptID))
	return nil
}

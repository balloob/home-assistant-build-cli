package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scriptCreateData   string
	scriptCreateFile   string
	scriptCreateFormat string
)

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
	scriptCreateCmd.Flags().StringVarP(&scriptCreateData, "data", "d", "", "Script configuration as JSON")
	scriptCreateCmd.Flags().StringVarP(&scriptCreateFile, "file", "f", "", "Path to config file")
	scriptCreateCmd.Flags().StringVar(&scriptCreateFormat, "format", "", "Input format (json, yaml)")
}

func runScriptCreate(cmd *cobra.Command, args []string) error {
	scriptID := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	config, err := input.ParseInput(scriptCreateData, scriptCreateFile, scriptCreateFormat)
	if err != nil {
		return err
	}

	if _, ok := config["alias"]; !ok {
		return fmt.Errorf("script must have an 'alias' field")
	}

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	result, err := restClient.Post(fmt.Sprintf("config/script/config/%s", scriptID), config)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Script %s created successfully.", scriptID))
	return nil
}

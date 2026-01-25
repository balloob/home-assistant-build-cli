package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperInputBooleanCreateIcon    string
	helperInputBooleanCreateInitial bool
)

var helperInputBooleanCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new input boolean helper",
	Long:  `Create a new input boolean (toggle) helper.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runHelperInputBooleanCreate,
}

func init() {
	helperInputBooleanParentCmd.AddCommand(helperInputBooleanCreateCmd)
	helperInputBooleanCreateCmd.Flags().StringVarP(&helperInputBooleanCreateIcon, "icon", "i", "", "Icon for the helper (e.g., mdi:toggle-switch)")
	helperInputBooleanCreateCmd.Flags().BoolVar(&helperInputBooleanCreateInitial, "initial", false, "Initial value (true/false)")
}

func runHelperInputBooleanCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
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

	params := map[string]interface{}{
		"name": name,
	}

	if helperInputBooleanCreateIcon != "" {
		params["icon"] = helperInputBooleanCreateIcon
	}

	if cmd.Flags().Changed("initial") {
		params["initial"] = helperInputBooleanCreateInitial
	}

	result, err := ws.HelperCreate("input_boolean", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input boolean '%s' created successfully.", name))
	return nil
}

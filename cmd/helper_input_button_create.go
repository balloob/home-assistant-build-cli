package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	helperInputButtonCreateIcon string
)

var helperInputButtonCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new input button helper",
	Long:  `Create a new input button helper that can be pressed.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runHelperInputButtonCreate,
}

func init() {
	helperInputButtonParentCmd.AddCommand(helperInputButtonCreateCmd)
	helperInputButtonCreateCmd.Flags().StringVarP(&helperInputButtonCreateIcon, "icon", "i", "", "Icon for the helper (e.g., mdi:button-pointer)")
}

func runHelperInputButtonCreate(cmd *cobra.Command, args []string) error {
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

	if helperInputButtonCreateIcon != "" {
		params["icon"] = helperInputButtonCreateIcon
	}

	result, err := ws.HelperCreate("input_button", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input button '%s' created successfully.", name))
	return nil
}

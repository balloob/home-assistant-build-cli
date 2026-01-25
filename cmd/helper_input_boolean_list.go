package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputBooleanListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all input boolean helpers",
	Long:  `List all input boolean (toggle) helpers.`,
	RunE:  runHelperInputBooleanList,
}

func init() {
	helperInputBooleanParentCmd.AddCommand(helperInputBooleanListCmd)
}

func runHelperInputBooleanList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("input_boolean")
	if err != nil {
		return err
	}

	client.PrintOutput(helpers, textMode, "")
	return nil
}

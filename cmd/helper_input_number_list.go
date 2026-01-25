package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputNumberListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all input number helpers",
	Long:  `List all input number helpers.`,
	RunE:  runHelperInputNumberList,
}

func init() {
	helperInputNumberParentCmd.AddCommand(helperInputNumberListCmd)
}

func runHelperInputNumberList(cmd *cobra.Command, args []string) error {
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

	helpers, err := ws.HelperList("input_number")
	if err != nil {
		return err
	}

	client.PrintOutput(helpers, textMode, "")
	return nil
}

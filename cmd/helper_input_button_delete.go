package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputButtonDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id_or_id>",
	Short: "Delete an input button helper",
	Long: `Delete an input button helper by entity ID or just the ID.

Examples:
  hab helper-input-button delete input_button.my_button
  hab helper-input-button delete my_button`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperInputButtonDelete,
}

func init() {
	helperInputButtonParentCmd.AddCommand(helperInputButtonDeleteCmd)
}

func runHelperInputButtonDelete(cmd *cobra.Command, args []string) error {
	id := args[0]

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

	err = ws.DeleteHelperByEntityOrEntryID(id, "input_button")
	if err != nil {
		return fmt.Errorf("failed to delete input button: %w", err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input button '%s' deleted successfully.", id))
	return nil
}

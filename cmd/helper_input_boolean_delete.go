package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputBooleanDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id_or_id>",
	Short: "Delete an input boolean helper",
	Long: `Delete an input boolean helper by entity ID or just the ID.

Examples:
  hab helper-input-boolean delete input_boolean.my_toggle
  hab helper-input-boolean delete my_toggle`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperInputBooleanDelete,
}

func init() {
	helperInputBooleanParentCmd.AddCommand(helperInputBooleanDeleteCmd)
}

func runHelperInputBooleanDelete(cmd *cobra.Command, args []string) error {
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

	err = ws.DeleteHelperByEntityOrEntryID(id, "input_boolean")
	if err != nil {
		return fmt.Errorf("failed to delete input boolean: %w", err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input boolean '%s' deleted successfully.", id))
	return nil
}

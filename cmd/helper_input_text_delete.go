package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputTextDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id_or_id>",
	Short: "Delete an input text helper",
	Long: `Delete an input text helper by entity ID or just the ID.

Examples:
  hab helper-input-text delete input_text.my_text
  hab helper-input-text delete my_text`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperInputTextDelete,
}

func init() {
	helperInputTextParentCmd.AddCommand(helperInputTextDeleteCmd)
}

func runHelperInputTextDelete(cmd *cobra.Command, args []string) error {
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

	err = ws.DeleteHelperByEntityOrEntryID(id, "input_text")
	if err != nil {
		return fmt.Errorf("failed to delete input text: %w", err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input text '%s' deleted successfully.", id))
	return nil
}

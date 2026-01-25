package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperInputNumberDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id_or_id>",
	Short: "Delete an input number helper",
	Long: `Delete an input number helper by entity ID or just the ID.

Examples:
  hab helper-input-number delete input_number.my_slider
  hab helper-input-number delete my_slider`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperInputNumberDelete,
}

func init() {
	helperInputNumberParentCmd.AddCommand(helperInputNumberDeleteCmd)
}

func runHelperInputNumberDelete(cmd *cobra.Command, args []string) error {
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

	err = ws.DeleteHelperByEntityOrEntryID(id, "input_number")
	if err != nil {
		return fmt.Errorf("failed to delete input number: %w", err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Input number '%s' deleted successfully.", id))
	return nil
}

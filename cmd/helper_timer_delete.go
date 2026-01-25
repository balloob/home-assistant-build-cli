package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperTimerDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id_or_id>",
	Short: "Delete a timer helper",
	Long: `Delete a timer helper by entity ID or just the ID.

Examples:
  hab helper-timer delete timer.my_timer
  hab helper-timer delete my_timer`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperTimerDelete,
}

func init() {
	helperTimerParentCmd.AddCommand(helperTimerDeleteCmd)
}

func runHelperTimerDelete(cmd *cobra.Command, args []string) error {
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

	err = ws.DeleteHelperByEntityOrEntryID(id, "timer")
	if err != nil {
		return fmt.Errorf("failed to delete timer: %w", err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Timer '%s' deleted successfully.", id))
	return nil
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperCounterDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id_or_id>",
	Short: "Delete a counter helper",
	Long: `Delete a counter helper by entity ID or just the ID.

Examples:
  hab helper-counter delete counter.my_counter
  hab helper-counter delete my_counter`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperCounterDelete,
}

func init() {
	helperCounterParentCmd.AddCommand(helperCounterDeleteCmd)
}

func runHelperCounterDelete(cmd *cobra.Command, args []string) error {
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

	err = ws.DeleteHelperByEntityOrEntryID(id, "counter")
	if err != nil {
		return fmt.Errorf("failed to delete counter: %w", err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Counter '%s' deleted successfully.", id))
	return nil
}

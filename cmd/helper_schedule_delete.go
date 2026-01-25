package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperScheduleDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id_or_id>",
	Short: "Delete a schedule helper",
	Long: `Delete a schedule helper by entity ID or just the ID.

Examples:
  hab helper-schedule delete schedule.my_schedule
  hab helper-schedule delete my_schedule`,
	Args: cobra.ExactArgs(1),
	RunE: runHelperScheduleDelete,
}

func init() {
	helperScheduleParentCmd.AddCommand(helperScheduleDeleteCmd)
}

func runHelperScheduleDelete(cmd *cobra.Command, args []string) error {
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

	err = ws.DeleteHelperByEntityOrEntryID(id, "schedule")
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	result := map[string]interface{}{
		"id":      id,
		"deleted": true,
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Schedule '%s' deleted successfully.", id))
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var notificationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List persistent notifications",
	Long:  `List all current persistent notifications in Home Assistant.`,
	Example: `  hab notification list
  hab notification list --json`,
	RunE: runNotificationList,
}

func init() {
	notificationCmd.AddCommand(notificationListCmd)
}

func runNotificationList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	notifications, err := ws.NotificationList()
	if err != nil {
		return err
	}

	if len(notifications) == 0 {
		output.PrintOutput([]interface{}{}, textMode, "No notifications.")
		return nil
	}

	output.PrintOutput(notifications, textMode, "")
	return nil
}

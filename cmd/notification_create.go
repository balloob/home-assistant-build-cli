package cmd

import (
	"github.com/spf13/cobra"
)

var (
	notificationCreateTitle          string
	notificationCreateNotificationID string
)

var notificationCreateCmd = &cobra.Command{
	Use:   "create <message>",
	Short: "Create a persistent notification",
	Long:  `Create a persistent notification that appears in the Home Assistant sidebar.`,
	Example: `  hab notification create "Backup completed successfully"
  hab notification create "Motion detected in garage" --title "Security Alert"
  hab notification create "Update available" --title "System" --notification-id update_notice`,
	Args: cobra.ExactArgs(1),
	RunE: runNotificationCreate,
}

func init() {
	notificationCmd.AddCommand(notificationCreateCmd)
	notificationCreateCmd.Flags().StringVar(&notificationCreateTitle, "title", "", "Title for the notification")
	notificationCreateCmd.Flags().StringVar(&notificationCreateNotificationID, "notification-id", "", "Custom notification ID (allows updating an existing notification)")
}

func runNotificationCreate(cmd *cobra.Command, args []string) error {
	message := args[0]

	data := map[string]interface{}{
		"message": message,
	}
	if notificationCreateTitle != "" {
		data["title"] = notificationCreateTitle
	}
	if notificationCreateNotificationID != "" {
		data["notification_id"] = notificationCreateNotificationID
	}

	return callServiceAction("persistent_notification", "create", "Notification created.", data)
}

package cmd

import (
	"github.com/spf13/cobra"
)

var notificationDismissCmd = &cobra.Command{
	Use:   "dismiss <notification_id>",
	Short: "Dismiss a persistent notification",
	Long:  `Dismiss a persistent notification from the Home Assistant sidebar.`,
	Example: `  hab notification dismiss update_notice
  hab notification dismiss 12345678abcd`,
	Args: cobra.ExactArgs(1),
	RunE: runNotificationDismiss,
}

func init() {
	notificationCmd.AddCommand(notificationDismissCmd)
	annotateServiceMutationCommand(notificationDismissCmd, "write", "notification", []string{"args"})
}

func runNotificationDismiss(cmd *cobra.Command, args []string) error {
	notificationID := args[0]

	return callServiceAction("dismiss", "notification", "persistent_notification", "dismiss", "Notification dismissed.", map[string]interface{}{
		"notification_id": notificationID,
	})
}

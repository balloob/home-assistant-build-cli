package cmd

import (
	"github.com/spf13/cobra"
)

var calendarDeleteRecurrenceRange string

var calendarDeleteCmd = &cobra.Command{
	Use:   "delete <entity_id> <uid>",
	Short: "Delete a calendar event",
	Long:  `Delete an event from a Home Assistant calendar by its uid.`,
	Example: `  hab calendar delete calendar.personal abc123def456
  hab calendar delete calendar.personal abc123def456 --recurrence-range THISEVENT
  hab calendar delete calendar.personal abc123def456 --recurrence-range THISANDFUTURE`,
	Args: cobra.ExactArgs(2),
	RunE: runCalendarDelete,
}

func init() {
	calendarCmd.AddCommand(calendarDeleteCmd)
	calendarDeleteCmd.Flags().StringVar(&calendarDeleteRecurrenceRange, "recurrence-range", "", "For recurring events: THISEVENT or THISANDFUTURE")
}

func runCalendarDelete(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "calendar")
	uid := args[1]

	data := map[string]interface{}{
		"entity_id": entityID,
		"uid":       uid,
	}
	if calendarDeleteRecurrenceRange != "" {
		data["recurrence_range"] = calendarDeleteRecurrenceRange
	}

	return callServiceAction("calendar", "delete_event", "Event deleted.", data)
}

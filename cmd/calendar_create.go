package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	calendarCreateSummary     string
	calendarCreateStart       string
	calendarCreateEnd         string
	calendarCreateDescription string
	calendarCreateLocation    string
	calendarCreateAllDay      bool
)

var calendarCreateCmd = &cobra.Command{
	Use:   "create <entity_id>",
	Short: "Create a calendar event",
	Long:  `Create a new event on a Home Assistant calendar entity.`,
	Example: `  hab calendar create calendar.personal --summary "Team meeting" --start 2026-04-01T10:00:00 --end 2026-04-01T11:00:00
  hab calendar create calendar.personal --summary "Holiday" --start 2026-12-25 --end 2026-12-26 --all-day
  hab calendar create calendar.work --summary "Conference" --start 2026-05-01T09:00:00 --end 2026-05-01T17:00:00 --location "Oslo" --description "Annual tech conference"`,
	Args: cobra.ExactArgs(1),
	RunE: runCalendarCreate,
}

func init() {
	calendarCmd.AddCommand(calendarCreateCmd)
	calendarCreateCmd.Flags().StringVar(&calendarCreateSummary, "summary", "", "Event title/summary (required)")
	calendarCreateCmd.Flags().StringVar(&calendarCreateStart, "start", "", "Event start time, ISO 8601 date or datetime (required)")
	calendarCreateCmd.Flags().StringVar(&calendarCreateEnd, "end", "", "Event end time, ISO 8601 date or datetime (required)")
	calendarCreateCmd.Flags().StringVar(&calendarCreateDescription, "description", "", "Event description")
	calendarCreateCmd.Flags().StringVar(&calendarCreateLocation, "location", "", "Event location")
	calendarCreateCmd.Flags().BoolVar(&calendarCreateAllDay, "all-day", false, "Create as an all-day event (start/end as YYYY-MM-DD)")
	_ = calendarCreateCmd.MarkFlagRequired("summary")
	_ = calendarCreateCmd.MarkFlagRequired("start")
	_ = calendarCreateCmd.MarkFlagRequired("end")
}

func runCalendarCreate(cmd *cobra.Command, args []string) error {
	entityID := ensureDomainPrefix(args[0], "calendar")

	data := map[string]interface{}{
		"entity_id": entityID,
		"summary":   calendarCreateSummary,
	}

	// HA uses dtstart/dtend for all-day (date-only) vs start_date_time/end_date_time
	if calendarCreateAllDay {
		data["start_date"] = calendarCreateStart
		data["end_date"] = calendarCreateEnd
	} else {
		if len(calendarCreateStart) <= 10 {
			return fmt.Errorf("--start must include a time component for non-all-day events (use --all-day for date-only)")
		}
		data["start_date_time"] = calendarCreateStart
		data["end_date_time"] = calendarCreateEnd
	}

	if calendarCreateDescription != "" {
		data["description"] = calendarCreateDescription
	}
	if calendarCreateLocation != "" {
		data["location"] = calendarCreateLocation
	}

	return callServiceAction("calendar", "create_event", "Event created.", data)
}

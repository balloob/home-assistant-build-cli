package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var (
	calendarListStart string
	calendarListEnd   string
)

var calendarListCmd = &cobra.Command{
	Use:   "list <entity_id>",
	Short: "List calendar events",
	Long:  `List events from a calendar entity.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCalendarList,
}

func init() {
	calendarCmd.AddCommand(calendarListCmd)
	calendarListCmd.Flags().StringVarP(&calendarListStart, "start", "s", "", "Start time (ISO format)")
	calendarListCmd.Flags().StringVarP(&calendarListEnd, "end", "e", "", "End time (ISO format)")
}

func runCalendarList(cmd *cobra.Command, args []string) error {
	entityID := args[0]
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	// Build endpoint: /api/calendars/<entity_id>?start=...&end=...
	// The HA REST API requires start and end query parameters.
	// Default to current time through the next 7 days when not specified.
	now := time.Now().UTC()
	start := calendarListStart
	if start == "" {
		start = now.Format(time.RFC3339)
	}
	end := calendarListEnd
	if end == "" {
		end = now.Add(7 * 24 * time.Hour).Format(time.RFC3339)
	}

	params := url.Values{}
	params.Set("start", start)
	params.Set("end", end)

	endpoint := "calendars/" + entityID + "?" + params.Encode()

	result, err := restClient.Get(endpoint)
	if err != nil {
		return err
	}

	if result == nil {
		fmt.Println("No events found.")
		return nil
	}

	// The REST API returns a plain array of events, but the original WebSocket API
	// returned {"events": [...]}.  Wrap the result to preserve the expected structure.
	wrapped := map[string]interface{}{
		"events": result,
	}

	client.PrintOutput(wrapped, textMode, "")
	return nil
}

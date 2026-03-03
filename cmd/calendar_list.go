package cmd

import (
	"fmt"
	"net/url"

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
	// The HA REST API uses /api/calendars/<entity_id> with optional time range params.
	params := url.Values{}
	if calendarListStart != "" {
		params.Set("start", calendarListStart)
	}
	if calendarListEnd != "" {
		params.Set("end", calendarListEnd)
	}

	endpoint := "calendars/" + entityID
	if len(params) > 0 {
		endpoint = endpoint + "?" + params.Encode()
	}

	result, err := restClient.Get(endpoint)
	if err != nil {
		return err
	}

	if result == nil {
		fmt.Println("No events found.")
		return nil
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

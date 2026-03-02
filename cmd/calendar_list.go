package cmd

import (
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

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{
		"entity_id": entityID,
	}
	if calendarListStart != "" {
		params["start"] = calendarListStart
	}
	if calendarListEnd != "" {
		params["end"] = calendarListEnd
	}

	result, err := ws.SendCommand("calendar/event/list", params)
	if err != nil {
		return err
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

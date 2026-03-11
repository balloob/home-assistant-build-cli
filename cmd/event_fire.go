package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var eventFireInput InputFlags

var eventFireCmd = &cobra.Command{
	Use:   "fire <event_type>",
	Short: "Fire an event",
	Long:  `Fire a custom event on the Home Assistant event bus.`,
	Example: `  hab event fire my_custom_event
  hab event fire my_custom_event --data '{"device_id": "abc123", "action": "triggered"}'
  hab event fire my_custom_event --file event_data.yaml
  hab event fire call_service --data '{"domain": "light", "service": "turn_on"}'`,
	Args: cobra.ExactArgs(1),
	RunE: runEventFire,
}

func init() {
	eventCmd.AddCommand(eventFireCmd)
	eventFireInput.Register(eventFireCmd)
}

func runEventFire(cmd *cobra.Command, args []string) error {
	eventType := args[0]
	textMode := getTextMode()

	var eventData map[string]interface{}
	if eventFireInput.Data != "" || eventFireInput.File != "" {
		var err error
		eventData, err = eventFireInput.Parse()
		if err != nil {
			return err
		}
	}

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	if err := restClient.FireEvent(eventType, eventData); err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Event '%s' fired.", eventType))
	return nil
}

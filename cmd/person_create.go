package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	personCreateDeviceTrackers []string
	personCreateUserID         string
)

var personCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new person",
	Long:  `Create a new person entry for presence tracking.`,
	Example: `  hab person create "Alice"
  hab person create "Bob" --device-trackers device_tracker.bobs_phone
  hab person create "Carol" --device-trackers device_tracker.carol_phone,device_tracker.carol_tablet --user-id 12345`,
	Args: cobra.ExactArgs(1),
	RunE: runPersonCreate,
}

func init() {
	personCmd.AddCommand(personCreateCmd)
	personCreateCmd.Flags().StringSliceVar(&personCreateDeviceTrackers, "device-trackers", nil, "Device tracker entity IDs (comma-separated)")
	personCreateCmd.Flags().StringVar(&personCreateUserID, "user-id", "", "HA user account ID to link to this person")
}

func runPersonCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{}
	if len(personCreateDeviceTrackers) > 0 {
		params["device_trackers"] = personCreateDeviceTrackers
	}
	if personCreateUserID != "" {
		params["user_id"] = personCreateUserID
	}

	result, err := ws.PersonRegistryCreate(name, params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Person '%s' created.", name))
	return nil
}

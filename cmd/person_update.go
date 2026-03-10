package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	personUpdateName           string
	personUpdateDeviceTrackers []string
	personUpdateUserID         string
)

var personUpdateCmd = &cobra.Command{
	Use:   "update <person_id>",
	Short: "Update a person",
	Long:  `Update an existing person entry.`,
	Example: `  hab person update ada6789 --name "Ada Lovelace"
  hab person update ada6789 --device-trackers device_tracker.ada_phone,device_tracker.ada_watch`,
	Args: cobra.ExactArgs(1),
	RunE: runPersonUpdate,
}

func init() {
	personCmd.AddCommand(personUpdateCmd)
	personUpdateCmd.Flags().StringVar(&personUpdateName, "name", "", "New name for the person")
	personUpdateCmd.Flags().StringSliceVar(&personUpdateDeviceTrackers, "device-trackers", nil, "Device tracker entity IDs (comma-separated)")
	personUpdateCmd.Flags().StringVar(&personUpdateUserID, "user-id", "", "HA user account ID to link")
}

func runPersonUpdate(cmd *cobra.Command, args []string) error {
	personID := args[0]
	textMode := getTextMode()

	params := map[string]interface{}{}
	if personUpdateName != "" {
		params["name"] = personUpdateName
	}
	if cmd.Flags().Changed("device-trackers") {
		params["device_trackers"] = personUpdateDeviceTrackers
	}
	if personUpdateUserID != "" {
		params["user_id"] = personUpdateUserID
	}

	if len(params) == 0 {
		return fmt.Errorf("no update parameters provided")
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.PersonRegistryUpdate(personID, params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Person '%s' updated.", personID))
	return nil
}

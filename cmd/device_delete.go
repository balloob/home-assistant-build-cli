package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var deviceDeleteForce bool

var deviceDeleteCmd = &cobra.Command{
	Use:   "delete <device_id>",
	Short: "Delete a device",
	Long:  `Delete a device from Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDeviceDelete,
}

func init() {
	deviceCmd.AddCommand(deviceDeleteCmd)
	deviceDeleteCmd.Flags().BoolVarP(&deviceDeleteForce, "force", "f", false, "Skip confirmation")
}

func runDeviceDelete(cmd *cobra.Command, args []string) error {
	deviceID := args[0]
	textMode := getTextMode()

	if err := confirmAction(deviceDeleteForce, fmt.Sprintf("Delete device %s? This will also remove all its entities.", deviceID), "delete device"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	_, err = ws.SendCommand("config/device_registry/remove_config_entry", map[string]interface{}{
		"device_id": deviceID,
	})
	if err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Device '%s' deleted.", deviceID), output.EnvelopeContext{Operation: "delete", ResourceType: "device"})
	return nil
}

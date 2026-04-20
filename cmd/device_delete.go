package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	deviceDeleteForce  bool
	deviceDeletePlan   bool
	deviceDeleteDryRun bool
)

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
	registerMutationPlanFlags(deviceDeleteCmd, &deviceDeletePlan, &deviceDeleteDryRun)
	mergeSchemaAnnotation(deviceDeleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "device",
		InputSources: []string{"args", "flags"},
	})
}

func runDeviceDelete(cmd *cobra.Command, args []string) error {
	deviceID := args[0]
	textMode := getTextMode()

	if planRequested(deviceDeletePlan, deviceDeleteDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":  "device",
				"device_id": deviceID,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send config/device_registry/remove_config_entry for the selected device ID.",
				"Return deletion confirmation.",
			},
			Risks: []string{
				"Deleting a device removes its registry entry and may remove linked entities.",
			},
			RequiresConfirmation: !deviceDeleteForce,
			VerificationCommands: []string{
				"hab device list --json",
			},
		}, textMode, "device")
		return nil
	}

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

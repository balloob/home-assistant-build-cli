package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	zoneDeleteForce bool
	zoneDeleteID    string
	zoneDeletePlan  bool
	zoneDeleteDry   bool
)

var zoneDeleteCmd = &cobra.Command{
	Use:   "delete [zone_id]",
	Short: "Delete a zone",
	Long:  `Delete a zone from Home Assistant.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runZoneDelete,
}

func init() {
	zoneCmd.AddCommand(zoneDeleteCmd)
	zoneDeleteCmd.Flags().StringVar(&zoneDeleteID, "zone", "", "Zone ID to delete")
	zoneDeleteCmd.Flags().StringVar(&zoneDeleteID, "zone-id", "", "Alias for --zone")
	zoneDeleteCmd.Flags().BoolVarP(&zoneDeleteForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(zoneDeleteCmd, &zoneDeletePlan, &zoneDeleteDry)
	mergeSchemaAnnotation(zoneDeleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "zone",
		InputSources: []string{"args", "flags"},
	})
}

func runZoneDelete(cmd *cobra.Command, args []string) error {
	zoneID, err := resolveArg(zoneDeleteID, args, 0, "zone ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	if planRequested(zoneDeletePlan, zoneDeleteDry) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource": "zone",
				"zone_id":  zoneID,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send zone delete command for the selected zone ID.",
				"Return deletion confirmation.",
			},
			Risks: []string{
				"Deleting a zone may impact location-based automations.",
			},
			RequiresConfirmation: !zoneDeleteForce,
			VerificationCommands: []string{
				"hab zone list --json",
			},
		}, textMode, "zone")
		return nil
	}

	if err := confirmAction(zoneDeleteForce, fmt.Sprintf("Delete zone %s?", zoneID), "delete zone"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.ZoneDelete(zoneID); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Zone '%s' deleted.", zoneID), output.EnvelopeContext{Operation: "delete", ResourceType: "zone"})
	return nil
}

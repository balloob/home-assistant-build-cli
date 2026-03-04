package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	zoneDeleteForce bool
	zoneDeleteID    string
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
	zoneDeleteCmd.Flags().BoolVarP(&zoneDeleteForce, "force", "f", false, "Skip confirmation")
}

func runZoneDelete(cmd *cobra.Command, args []string) error {
	zoneID, err := resolveArg(zoneDeleteID, args, 0, "zone ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	if !confirmAction(zoneDeleteForce, textMode, fmt.Sprintf("Delete zone %s?", zoneID)) {
		fmt.Println("Cancelled.")
		return nil
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.ZoneDelete(zoneID); err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Zone '%s' deleted.", zoneID))
	return nil
}

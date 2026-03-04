package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var zoneListFlags *ListFlags

var zoneListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones",
	Long:  `List all zones in Home Assistant.`,
	RunE:  runZoneList,
}

func init() {
	zoneCmd.AddCommand(zoneListCmd)
	zoneListFlags = RegisterListFlags(zoneListCmd, "id")
}

func runZoneList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	zones, err := ws.ZoneList()
	if err != nil {
		return err
	}

	if zoneListFlags.RenderCount(len(zones), textMode) {
		return nil
	}
	zones = zoneListFlags.ApplyLimit(zones)
	if zoneListFlags.RenderBrief(zones, textMode, "id", "name") {
		return nil
	}

	output.PrintOutput(zones, textMode, "")
	return nil
}

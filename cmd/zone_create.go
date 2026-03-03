package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	zoneCreateLatitude  float64
	zoneCreateLongitude float64
	zoneCreateRadius    float64
	zoneCreateIcon      string
	zoneCreatePassive   bool
)

var zoneCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new zone",
	Long:  `Create a new zone in Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runZoneCreate,
}

func init() {
	zoneCmd.AddCommand(zoneCreateCmd)
	zoneCreateCmd.Flags().Float64Var(&zoneCreateLatitude, "latitude", 0, "Latitude of the zone center (required)")
	zoneCreateCmd.Flags().Float64Var(&zoneCreateLongitude, "longitude", 0, "Longitude of the zone center (required)")
	zoneCreateCmd.Flags().Float64Var(&zoneCreateRadius, "radius", 100, "Radius of the zone in meters")
	zoneCreateCmd.Flags().StringVar(&zoneCreateIcon, "icon", "", "Icon for the zone")
	zoneCreateCmd.Flags().BoolVar(&zoneCreatePassive, "passive", false, "Zone is passive (only for automations)")
	zoneCreateCmd.MarkFlagRequired("latitude")
	zoneCreateCmd.MarkFlagRequired("longitude")
}

func runZoneCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := make(map[string]interface{})
	if zoneCreateIcon != "" {
		params["icon"] = zoneCreateIcon
	}
	if zoneCreatePassive {
		params["passive"] = true
	}

	result, err := ws.ZoneCreate(name, zoneCreateLatitude, zoneCreateLongitude, zoneCreateRadius, params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Zone '%s' created.", name))
	return nil
}

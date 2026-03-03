package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var (
	areaUpdateName  string
	areaUpdateFloor string
	areaUpdateIcon  string
)

var areaUpdateCmd = &cobra.Command{
	Use:   "update <area_id>",
	Short: "Update an area",
	Long:  `Update an existing area.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAreaUpdate,
}

func init() {
	areaCmd.AddCommand(areaUpdateCmd)
	areaUpdateCmd.Flags().StringVar(&areaUpdateName, "name", "", "New name for the area")
	areaUpdateCmd.Flags().StringVar(&areaUpdateFloor, "floor", "", "Floor ID to assign")
	areaUpdateCmd.Flags().StringVar(&areaUpdateIcon, "icon", "", "Icon for the area")
}

func runAreaUpdate(cmd *cobra.Command, args []string) error {
	areaID := args[0]
	textMode := getTextMode()

	params := make(map[string]interface{})
	if areaUpdateName != "" {
		params["name"] = areaUpdateName
	}
	if areaUpdateFloor != "" {
		params["floor_id"] = areaUpdateFloor
	}
	if areaUpdateIcon != "" {
		params["icon"] = areaUpdateIcon
	}

	if len(params) == 0 {
		return fmt.Errorf("no update parameters provided")
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.AreaRegistryUpdate(areaID, params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Area '%s' updated.", areaID))
	return nil
}

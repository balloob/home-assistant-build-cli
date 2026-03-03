package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	floorUpdateName  string
	floorUpdateIcon  string
	floorUpdateLevel int
)

var floorUpdateCmd = &cobra.Command{
	Use:   "update <floor_id>",
	Short: "Update a floor",
	Long:  `Update an existing floor.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runFloorUpdate,
}

func init() {
	floorCmd.AddCommand(floorUpdateCmd)
	floorUpdateCmd.Flags().StringVar(&floorUpdateName, "name", "", "New name for the floor")
	floorUpdateCmd.Flags().StringVar(&floorUpdateIcon, "icon", "", "Icon for the floor")
	floorUpdateCmd.Flags().IntVar(&floorUpdateLevel, "level", 0, "Floor level")
}

func runFloorUpdate(cmd *cobra.Command, args []string) error {
	floorID := args[0]
	textMode := getTextMode()

	params := make(map[string]interface{})
	if floorUpdateName != "" {
		params["name"] = floorUpdateName
	}
	if floorUpdateIcon != "" {
		params["icon"] = floorUpdateIcon
	}
	if cmd.Flags().Changed("level") {
		params["level"] = floorUpdateLevel
	}

	if len(params) == 0 {
		return fmt.Errorf("no update parameters provided")
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.FloorRegistryUpdate(floorID, params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Floor '%s' updated.", floorID))
	return nil
}

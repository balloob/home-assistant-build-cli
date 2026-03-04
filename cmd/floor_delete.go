package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var floorDeleteForce bool

var floorDeleteCmd = &cobra.Command{
	Use:   "delete <floor_id>",
	Short: "Delete a floor",
	Long:  `Delete a floor from Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runFloorDelete,
}

func init() {
	floorCmd.AddCommand(floorDeleteCmd)
	floorDeleteCmd.Flags().BoolVarP(&floorDeleteForce, "force", "f", false, "Skip confirmation")
}

func runFloorDelete(cmd *cobra.Command, args []string) error {
	floorID := args[0]
	textMode := getTextMode()

	if !confirmAction(floorDeleteForce, textMode, fmt.Sprintf("Delete floor %s?", floorID)) {
		fmt.Println("Cancelled.")
		return nil
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.FloorRegistryDelete(floorID); err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Floor '%s' deleted.", floorID))
	return nil
}

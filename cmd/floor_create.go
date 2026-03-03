package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	floorCreateIcon  string
	floorCreateLevel int
)

var floorCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new floor",
	Long:  `Create a new floor in Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runFloorCreate,
}

func init() {
	floorCmd.AddCommand(floorCreateCmd)
	floorCreateCmd.Flags().StringVar(&floorCreateIcon, "icon", "", "Icon for the floor")
	floorCreateCmd.Flags().IntVar(&floorCreateLevel, "level", 0, "Floor level (0 = ground)")
}

func runFloorCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := make(map[string]interface{})
	if floorCreateIcon != "" {
		params["icon"] = floorCreateIcon
	}
	if cmd.Flags().Changed("level") {
		params["level"] = floorCreateLevel
	}

	result, err := ws.FloorRegistryCreate(name, params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Floor '%s' created.", name))
	return nil
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	areaCreateFloor string
	areaCreateIcon  string
)

var areaCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new area",
	Long:  `Create a new area in Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAreaCreate,
}

func init() {
	areaCmd.AddCommand(areaCreateCmd)
	areaCreateCmd.Flags().StringVar(&areaCreateFloor, "floor", "", "Floor ID to assign")
	areaCreateCmd.Flags().StringVar(&areaCreateIcon, "icon", "", "Icon for the area")
}

func runAreaCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := make(map[string]interface{})
	if areaCreateFloor != "" {
		params["floor_id"] = areaCreateFloor
	}
	if areaCreateIcon != "" {
		params["icon"] = areaCreateIcon
	}

	result, err := ws.AreaRegistryCreate(name, params)
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Area '%s' created.", name))
	return nil
}

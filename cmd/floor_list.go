package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var floorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all floors",
	Long:  `List all floors in Home Assistant.`,
	RunE:  runFloorList,
}

var (
	floorListID    string
	floorListCount bool
	floorListBrief bool
	floorListLimit int
)

func init() {
	floorCmd.AddCommand(floorListCmd)
	floorListCmd.Flags().StringVar(&floorListID, "floor-id", "", "Filter by floor ID")
	floorListCmd.Flags().BoolVarP(&floorListCount, "count", "c", false, "Return only the count of items")
	floorListCmd.Flags().BoolVarP(&floorListBrief, "brief", "b", false, "Return minimal fields (floor_id and name only)")
	floorListCmd.Flags().IntVarP(&floorListLimit, "limit", "n", 0, "Limit results to N items")
}

func runFloorList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	floors, err := ws.FloorRegistryList()
	if err != nil {
		return err
	}

	// Apply floor ID filter
	if floorListID != "" {
		var filtered []interface{}
		for _, f := range floors {
			if floor, ok := f.(map[string]interface{}); ok {
				floorID, _ := floor["floor_id"].(string)
				if floorID == floorListID {
					filtered = append(filtered, f)
				}
			}
		}
		floors = filtered
	}

	// Handle count mode
	if floorListCount {
		if textMode {
			fmt.Printf("Count: %d\n", len(floors))
		} else {
			output.PrintOutput(map[string]interface{}{"count": len(floors)}, false, "")
		}
		return nil
	}

	// Apply limit
	if floorListLimit > 0 && len(floors) > floorListLimit {
		floors = floors[:floorListLimit]
	}

	// Handle brief mode
	if floorListBrief {
		if textMode {
			for _, f := range floors {
				if floor, ok := f.(map[string]interface{}); ok {
					name, _ := floor["name"].(string)
					floorID, _ := floor["floor_id"].(string)
					fmt.Printf("%s (%s)\n", name, floorID)
				}
			}
		} else {
			var brief []map[string]interface{}
			for _, f := range floors {
				if floor, ok := f.(map[string]interface{}); ok {
					brief = append(brief, map[string]interface{}{
						"floor_id": floor["floor_id"],
						"name":     floor["name"],
					})
				}
			}
			output.PrintOutput(brief, false, "")
		}
		return nil
	}

	// Full output
	if textMode {
		if len(floors) == 0 {
			fmt.Println("No floors.")
			return nil
		}
		for _, f := range floors {
			if floor, ok := f.(map[string]interface{}); ok {
				name, _ := floor["name"].(string)
				floorID, _ := floor["floor_id"].(string)
				level, hasLevel := floor["level"].(float64)

				if hasLevel {
					fmt.Printf("%s (%s): level %.0f\n", name, floorID, level)
				} else {
					fmt.Printf("%s (%s)\n", name, floorID)
				}
			}
		}
	} else {
		output.PrintOutput(floors, false, "")
	}
	return nil
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/client"
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

	if !floorDeleteForce && !textMode {
		fmt.Printf("Delete floor %s? [y/N]: ", floorID)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.FloorRegistryDelete(floorID); err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Floor '%s' deleted.", floorID))
	return nil
}

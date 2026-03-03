package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var areaDeleteForce bool

var areaDeleteCmd = &cobra.Command{
	Use:   "delete <area_id>",
	Short: "Delete an area",
	Long:  `Delete an area from Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAreaDelete,
}

func init() {
	areaCmd.AddCommand(areaDeleteCmd)
	areaDeleteCmd.Flags().BoolVarP(&areaDeleteForce, "force", "f", false, "Skip confirmation")
}

func runAreaDelete(cmd *cobra.Command, args []string) error {
	areaID := args[0]
	textMode := getTextMode()

	if !areaDeleteForce && !textMode {
		fmt.Printf("Delete area %s? [y/N]: ", areaID)
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

	if err := ws.AreaRegistryDelete(areaID); err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Area '%s' deleted.", areaID))
	return nil
}

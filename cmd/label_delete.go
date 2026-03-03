package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var labelDeleteForce bool

var labelDeleteCmd = &cobra.Command{
	Use:   "delete <label_id>",
	Short: "Delete a label",
	Long:  `Delete a label from Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runLabelDelete,
}

func init() {
	labelCmd.AddCommand(labelDeleteCmd)
	labelDeleteCmd.Flags().BoolVarP(&labelDeleteForce, "force", "f", false, "Skip confirmation")
}

func runLabelDelete(cmd *cobra.Command, args []string) error {
	labelID := args[0]
	textMode := getTextMode()

	if !labelDeleteForce && !textMode {
		fmt.Printf("Delete label %s? [y/N]: ", labelID)
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

	if err := ws.LabelRegistryDelete(labelID); err != nil {
		return err
	}

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Label '%s' deleted.", labelID))
	return nil
}

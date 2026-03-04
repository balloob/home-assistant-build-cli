package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
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

	if !confirmAction(labelDeleteForce, textMode, fmt.Sprintf("Delete label %s?", labelID)) {
		fmt.Println("Cancelled.")
		return nil
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.LabelRegistryDelete(labelID); err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Label '%s' deleted.", labelID))
	return nil
}

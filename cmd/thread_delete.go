package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	threadDeleteForce     bool
	threadDeleteDatasetID string
)

var threadDeleteCmd = &cobra.Command{
	Use:   "delete [dataset_id]",
	Short: "Delete a Thread dataset",
	Long:  `Delete a Thread dataset.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runThreadDelete,
}

func init() {
	threadCmd.AddCommand(threadDeleteCmd)
	threadDeleteCmd.Flags().StringVar(&threadDeleteDatasetID, "dataset", "", "Thread dataset ID to delete")
	threadDeleteCmd.Flags().BoolVarP(&threadDeleteForce, "force", "f", false, "Skip confirmation")
}

func runThreadDelete(cmd *cobra.Command, args []string) error {
	datasetID, err := resolveArg(threadDeleteDatasetID, args, 0, "dataset ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	if !confirmAction(threadDeleteForce, textMode, fmt.Sprintf("Delete Thread dataset %s?", datasetID)) {
		fmt.Println("Cancelled.")
		return nil
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	_, err = ws.SendCommand("thread/delete_dataset", map[string]interface{}{
		"dataset_id": datasetID,
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Thread dataset %s deleted.", datasetID))
	return nil
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var threadSetPreferredDatasetID string

var threadSetPreferredCmd = &cobra.Command{
	Use:   "set-preferred [dataset_id]",
	Short: "Set a dataset as the preferred network",
	Long:  `Set a Thread dataset as the preferred network.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runThreadSetPreferred,
}

func init() {
	threadCmd.AddCommand(threadSetPreferredCmd)
	threadSetPreferredCmd.Flags().StringVar(&threadSetPreferredDatasetID, "dataset", "", "Thread dataset ID to set as preferred")
}

func runThreadSetPreferred(cmd *cobra.Command, args []string) error {
	datasetID := threadSetPreferredDatasetID
	if datasetID == "" && len(args) > 0 {
		datasetID = args[0]
	}
	if datasetID == "" {
		return fmt.Errorf("dataset ID is required (use --dataset flag or positional argument)")
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	_, err = ws.SendCommand("thread/set_preferred_dataset", map[string]interface{}{
		"dataset_id": datasetID,
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, fmt.Sprintf("Thread dataset %s set as preferred.", datasetID))
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var threadGetDatasetID string

var threadGetCmd = &cobra.Command{
	Use:   "get [dataset_id]",
	Short: "Get dataset details including TLV",
	Long:  `Get full details of a Thread dataset.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runThreadGet,
}

func init() {
	threadCmd.AddCommand(threadGetCmd)
	threadGetCmd.Flags().StringVar(&threadGetDatasetID, "dataset", "", "Thread dataset ID to get")
}

func runThreadGet(cmd *cobra.Command, args []string) error {
	datasetID, err := resolveArg(threadGetDatasetID, args, 0, "dataset ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("thread/get_dataset_tlv", map[string]interface{}{
		"dataset_id": datasetID,
	})
	if err != nil {
		return err
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var threadListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Thread datasets",
	Long:  `List all Thread network datasets.`,
	RunE:  runThreadList,
}

func init() {
	threadCmd.AddCommand(threadListCmd)
}

func runThreadList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("thread/list_datasets", nil)
	if err != nil {
		return err
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

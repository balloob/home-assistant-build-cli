package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var eventListCmd = &cobra.Command{
	Use:   "list",
	Short: "List event types",
	Long:  `List all registered event types and their current listener counts.`,
	Example: `  hab event list
  hab event list --json`,
	RunE: runEventList,
}

func init() {
	eventCmd.AddCommand(eventListCmd)
}

func runEventList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	events, err := restClient.GetEvents()
	if err != nil {
		return err
	}

	if len(events) == 0 {
		output.PrintOutput([]interface{}{}, textMode, "No events found.")
		return nil
	}

	output.PrintOutput(events, textMode, "")
	return nil
}

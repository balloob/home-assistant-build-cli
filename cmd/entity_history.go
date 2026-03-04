package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	entityHistoryStart string
	entityHistoryEnd   string
	entityHistoryID    string
)

var entityHistoryCmd = &cobra.Command{
	Use:   "history [entity_id]",
	Short: "Get state history",
	Long:  `Get the state history for an entity.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runEntityHistory,
}

func init() {
	entityCmd.AddCommand(entityHistoryCmd)
	entityHistoryCmd.Flags().StringVar(&entityHistoryID, "entity", "", "Entity ID to get history for")
	entityHistoryCmd.Flags().StringVarP(&entityHistoryStart, "start", "s", "", "Start time (ISO format)")
	entityHistoryCmd.Flags().StringVarP(&entityHistoryEnd, "end", "e", "", "End time (ISO format)")
}

func runEntityHistory(cmd *cobra.Command, args []string) error {
	entityID, err := resolveArg(entityHistoryID, args, 0, "entity ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	history, err := restClient.GetHistory(entityID, entityHistoryStart, entityHistoryEnd)
	if err != nil {
		return err
	}

	// Flatten the nested list
	if len(history) > 0 {
		output.PrintOutput(history[0], textMode, "")
	} else {
		output.PrintOutput([]interface{}{}, textMode, "")
	}
	return nil
}

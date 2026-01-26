package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	entityID := entityHistoryID
	if entityID == "" && len(args) > 0 {
		entityID = args[0]
	}
	if entityID == "" {
		return fmt.Errorf("entity ID is required (use --entity flag or positional argument)")
	}
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	history, err := restClient.GetHistory(entityID, entityHistoryStart, entityHistoryEnd)
	if err != nil {
		return err
	}

	// Flatten the nested list
	if len(history) > 0 {
		client.PrintOutput(history[0], textMode, "")
	} else {
		client.PrintOutput([]interface{}{}, textMode, "")
	}
	return nil
}

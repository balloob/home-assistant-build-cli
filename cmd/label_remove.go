package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	labelRemoveLabelID  string
	labelRemoveEntityID string
)

var labelRemoveCmd = &cobra.Command{
	Use:   "remove [label_id] [entity_id]",
	Short: "Remove label from entity",
	Long:  `Remove a label from an entity.`,
	Args:  cobra.MaximumNArgs(2),
	RunE:  runLabelRemove,
}

func init() {
	labelCmd.AddCommand(labelRemoveCmd)
	labelRemoveCmd.Flags().StringVar(&labelRemoveLabelID, "label", "", "Label ID to remove")
	labelRemoveCmd.Flags().StringVar(&labelRemoveEntityID, "entity", "", "Entity ID to remove the label from")
}

func runLabelRemove(cmd *cobra.Command, args []string) error {
	labelID := labelRemoveLabelID
	if labelID == "" && len(args) > 0 {
		labelID = args[0]
	}
	if labelID == "" {
		return fmt.Errorf("label ID is required (use --label flag or first positional argument)")
	}
	entityID := labelRemoveEntityID
	if entityID == "" && len(args) > 1 {
		entityID = args[1]
	}
	if entityID == "" {
		return fmt.Errorf("entity ID is required (use --entity flag or second positional argument)")
	}
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	// First get current entity labels
	entity, err := ws.EntityRegistryGet(entityID)
	if err != nil {
		return err
	}

	currentLabels, _ := entity["labels"].([]interface{})
	labels := make([]string, 0, len(currentLabels))
	found := false
	for _, l := range currentLabels {
		if ls, ok := l.(string); ok {
			if ls == labelID {
				found = true
				continue
			}
			labels = append(labels, ls)
		}
	}

	if !found {
		client.PrintSuccess(nil, textMode, fmt.Sprintf("Entity %s does not have label %s.", entityID, labelID))
		return nil
	}

	// Update entity with new labels
	result, err := ws.EntityRegistryUpdate(entityID, map[string]interface{}{
		"labels": labels,
	})
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Label %s removed from %s.", labelID, entityID))
	return nil
}

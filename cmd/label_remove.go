package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
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
	labelID, err := resolveArg(labelRemoveLabelID, args, 0, "label ID")
	if err != nil {
		return err
	}
	entityID, err := resolveArg(labelRemoveEntityID, args, 1, "entity ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
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
		output.PrintSuccess(nil, textMode, fmt.Sprintf("Entity %s does not have label %s.", entityID, labelID))
		return nil
	}

	// Update entity with new labels
	result, err := ws.EntityRegistryUpdate(entityID, map[string]interface{}{
		"labels": labels,
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Label %s removed from %s.", labelID, entityID))
	return nil
}

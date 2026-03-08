package cmd

import (
	"fmt"

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

	return modifyEntityLabels(entityID, labelID, func(labels []string) ([]string, string) {
		result := make([]string, 0, len(labels))
		found := false
		for _, l := range labels {
			if l == labelID {
				found = true
				continue
			}
			result = append(result, l)
		}
		if !found {
			return nil, fmt.Sprintf("Entity %s does not have label %s.", entityID, labelID)
		}
		return result, fmt.Sprintf("Label %s removed from %s.", labelID, entityID)
	})
}

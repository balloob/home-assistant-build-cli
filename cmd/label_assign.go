package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	labelAssignLabelID  string
	labelAssignEntityID string
)

var labelAssignCmd = &cobra.Command{
	Use:   "assign [label_id] [entity_id]",
	Short: "Assign label to entity",
	Long:  `Assign a label to an entity.`,
	Args:  cobra.MaximumNArgs(2),
	RunE:  runLabelAssign,
}

func init() {
	labelCmd.AddCommand(labelAssignCmd)
	labelAssignCmd.Flags().StringVar(&labelAssignLabelID, "label", "", "Label ID to assign")
	labelAssignCmd.Flags().StringVar(&labelAssignEntityID, "entity", "", "Entity ID to assign the label to")
}

func runLabelAssign(cmd *cobra.Command, args []string) error {
	labelID, err := resolveArg(labelAssignLabelID, args, 0, "label ID")
	if err != nil {
		return err
	}
	entityID, err := resolveArg(labelAssignEntityID, args, 1, "entity ID")
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
	labels := make([]string, 0, len(currentLabels)+1)
	for _, l := range currentLabels {
		if ls, ok := l.(string); ok {
			if ls == labelID {
				// Already has label
				output.PrintSuccess(nil, textMode, fmt.Sprintf("Entity %s already has label %s.", entityID, labelID))
				return nil
			}
			labels = append(labels, ls)
		}
	}
	labels = append(labels, labelID)

	// Update entity with new labels
	result, err := ws.EntityRegistryUpdate(entityID, map[string]interface{}{
		"labels": labels,
	})
	if err != nil {
		return err
	}

	output.PrintSuccess(result, textMode, fmt.Sprintf("Label %s assigned to %s.", labelID, entityID))
	return nil
}

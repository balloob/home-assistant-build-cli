package cmd

import (
	"fmt"

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

	return modifyEntityLabels(entityID, labelID, func(labels []string) ([]string, string) {
		for _, l := range labels {
			if l == labelID {
				return nil, fmt.Sprintf("Entity %s already has label %s.", entityID, labelID)
			}
		}
		return append(labels, labelID), fmt.Sprintf("Label %s assigned to %s.", labelID, entityID)
	})
}

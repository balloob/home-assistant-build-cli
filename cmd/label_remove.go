package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	labelRemoveLabelID  string
	labelRemoveEntityID string
	labelRemoveForce    bool
	labelRemovePlan     bool
	labelRemoveDryRun   bool
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
	labelRemoveCmd.Flags().StringVar(&labelRemoveLabelID, "label-id", "", "Alias for --label")
	labelRemoveCmd.Flags().StringVar(&labelRemoveEntityID, "entity", "", "Entity ID to remove the label from")
	labelRemoveCmd.Flags().StringVar(&labelRemoveEntityID, "entity-id", "", "Alias for --entity")
	labelRemoveCmd.Flags().BoolVarP(&labelRemoveForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(labelRemoveCmd, &labelRemovePlan, &labelRemoveDryRun)
	mergeSchemaAnnotation(labelRemoveCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "label_assignment",
		InputSources: []string{"args", "flags"},
	})
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

	if planRequested(labelRemovePlan, labelRemoveDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":  "label_assignment",
				"entity_id": entityID,
				"label_id":  labelID,
			},
			Steps: []string{
				"Fetch current labels for the target entity.",
				"Remove the selected label ID from the entity label list.",
				"Update entity registry labels.",
			},
			Risks: []string{
				"Removing the label changes entity organization and automations that rely on labels.",
			},
			RequiresConfirmation: !labelRemoveForce,
			VerificationCommands: []string{
				fmt.Sprintf("hab search related entity %s --json", entityID),
			},
		}, textMode, "label_assignment")
		return nil
	}

	if err := confirmAction(labelRemoveForce, fmt.Sprintf("Remove label %s from %s?", labelID, entityID), "remove label assignment"); err != nil {
		return err
	}

	return modifyEntityLabels(entityID, labelID, "remove", func(labels []string) ([]string, string) {
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

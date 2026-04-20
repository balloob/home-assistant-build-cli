package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	threadDeleteForce     bool
	threadDeleteDatasetID string
	threadDeletePlan      bool
	threadDeleteDryRun    bool
)

var threadDeleteCmd = &cobra.Command{
	Use:   "delete [dataset_id]",
	Short: "Delete a Thread dataset",
	Long:  `Delete a Thread dataset.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runThreadDelete,
}

func init() {
	threadCmd.AddCommand(threadDeleteCmd)
	threadDeleteCmd.Flags().StringVar(&threadDeleteDatasetID, "dataset", "", "Thread dataset ID to delete")
	threadDeleteCmd.Flags().StringVar(&threadDeleteDatasetID, "dataset-id", "", "Alias for --dataset")
	threadDeleteCmd.Flags().BoolVarP(&threadDeleteForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(threadDeleteCmd, &threadDeletePlan, &threadDeleteDryRun)
	mergeSchemaAnnotation(threadDeleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "thread_dataset",
		InputSources: []string{"args", "flags"},
	})
}

func runThreadDelete(cmd *cobra.Command, args []string) error {
	datasetID, err := resolveArg(threadDeleteDatasetID, args, 0, "dataset ID")
	if err != nil {
		return err
	}
	textMode := getTextMode()

	if planRequested(threadDeletePlan, threadDeleteDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":   "thread_dataset",
				"dataset_id": datasetID,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send thread/delete_dataset command with the selected dataset ID.",
				"Return deletion confirmation.",
			},
			Risks: []string{
				"Deleting a Thread dataset may disrupt border router network configuration.",
			},
			RequiresConfirmation: !threadDeleteForce,
			VerificationCommands: []string{
				"hab thread list --json",
			},
		}, textMode, "thread_dataset")
		return nil
	}

	if err := confirmAction(threadDeleteForce, fmt.Sprintf("Delete Thread dataset %s?", datasetID), "delete thread dataset"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	_, err = ws.SendCommand("thread/delete_dataset", map[string]interface{}{
		"dataset_id": datasetID,
	})
	if err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Thread dataset %s deleted.", datasetID), output.EnvelopeContext{Operation: "delete", ResourceType: "thread_dataset"})
	return nil
}

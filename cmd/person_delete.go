package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	personDeleteForce  bool
	personDeletePlan   bool
	personDeleteDryRun bool
)

var personDeleteCmd = &cobra.Command{
	Use:   "delete <person_id>",
	Short: "Delete a person",
	Long:  `Delete a person entry from Home Assistant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPersonDelete,
}

func init() {
	personCmd.AddCommand(personDeleteCmd)
	personDeleteCmd.Flags().BoolVarP(&personDeleteForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(personDeleteCmd, &personDeletePlan, &personDeleteDryRun)
	mergeSchemaAnnotation(personDeleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "person",
		InputSources: []string{"args", "flags"},
	})
}

func runPersonDelete(cmd *cobra.Command, args []string) error {
	personID := args[0]
	textMode := getTextMode()

	if planRequested(personDeletePlan, personDeleteDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":  "person",
				"person_id": personID,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send person registry delete for the selected person ID.",
				"Return deletion confirmation.",
			},
			Risks: []string{
				"Deleting a person removes person registry metadata and links.",
			},
			RequiresConfirmation: !personDeleteForce,
			VerificationCommands: []string{
				"hab person list --json",
			},
		}, textMode, "person")
		return nil
	}

	if err := confirmAction(personDeleteForce, fmt.Sprintf("Delete person %s?", personID), "delete person"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.PersonRegistryDelete(personID); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Person '%s' deleted.", personID), output.EnvelopeContext{Operation: "delete", ResourceType: "person"})
	return nil
}

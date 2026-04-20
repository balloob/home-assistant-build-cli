package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	blueprintDeleteDomainValue string
	blueprintDeleteForce       bool
	blueprintDeletePlan        bool
	blueprintDeleteDryRun      bool
)

var blueprintDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a blueprint",
	Long:  `Delete a blueprint by its path. Use --domain to specify the domain (default: automation).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBlueprintDelete,
}

func init() {
	blueprintCmd.AddCommand(blueprintDeleteCmd)
	blueprintDeleteCmd.Flags().StringVar(&blueprintDeleteDomainValue, "domain", "automation", "Domain of the blueprint (automation/script)")
	blueprintDeleteCmd.Flags().BoolVarP(&blueprintDeleteForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(blueprintDeleteCmd, &blueprintDeletePlan, &blueprintDeleteDryRun)
	mergeSchemaAnnotation(blueprintDeleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "blueprint",
		InputSources: []string{"args", "flags"},
	})
}

func runBlueprintDelete(cmd *cobra.Command, args []string) error {
	path := args[0]
	textMode := getTextMode()

	if planRequested(blueprintDeletePlan, blueprintDeleteDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource": "blueprint",
				"domain":   blueprintDeleteDomainValue,
				"path":     path,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send blueprint delete command with the selected domain and path.",
				"Return deletion confirmation.",
			},
			Risks: []string{
				"Deleting a blueprint removes reusable automation/script templates from Home Assistant.",
			},
			RequiresConfirmation: !blueprintDeleteForce,
			VerificationCommands: []string{
				fmt.Sprintf("hab blueprint list %s --json", blueprintDeleteDomainValue),
			},
		}, textMode, "blueprint")
		return nil
	}

	if err := confirmAction(blueprintDeleteForce, fmt.Sprintf("Delete blueprint %s?", path), "delete blueprint"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("blueprint/delete", map[string]interface{}{
		"domain": blueprintDeleteDomainValue,
		"path":   path,
	})
	if err != nil {
		return err
	}

	output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("Blueprint %s deleted successfully.", path), output.EnvelopeContext{Operation: "delete", ResourceType: "blueprint"})
	return nil
}

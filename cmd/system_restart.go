package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	restartForce  bool
	restartPlan   bool
	restartDryRun bool
)

var systemRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Home Assistant",
	Long:  `Restart the Home Assistant core.`,
	RunE:  runSystemRestart,
}

func init() {
	systemCmd.AddCommand(systemRestartCmd)
	systemRestartCmd.Flags().BoolVarP(&restartForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(systemRestartCmd, &restartPlan, &restartDryRun)
	mergeSchemaAnnotation(systemRestartCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "rest"},
		ResourceType: "system",
		InputSources: []string{"flags"},
	})
}

func runSystemRestart(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	if planRequested(restartPlan, restartDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource": "system",
				"action":   "restart",
			},
			Steps: []string{
				"Call Home Assistant restart endpoint.",
				"Restart Home Assistant core.",
			},
			Risks: []string{
				"Restart interrupts active automations, integrations, and API availability briefly.",
			},
			RequiresConfirmation: !restartForce,
			VerificationCommands: []string{
				"hab system health --json",
			},
		}, textMode, "system")
		return nil
	}

	if err := confirmAction(restartForce, "This will restart Home Assistant. Continue?", "restart system"); err != nil {
		return err
	}

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	if err := restClient.Restart(); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, "Restart initiated.", output.EnvelopeContext{Operation: "restart", ResourceType: "system"})
	return nil
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	networkConfigureAdapters []string
	networkConfigureApply    bool
	networkConfigureForce    bool
	networkConfigurePlan     bool
	networkConfigureDryRun   bool
)

var networkConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure network adapters",
	Long: `Configure the set of network adapters Home Assistant should use.

This command is potentially disruptive and requires --apply.`,
	RunE: runNetworkConfigure,
}

func init() {
	networkCmd.AddCommand(networkConfigureCmd)
	networkConfigureCmd.Flags().StringSliceVar(&networkConfigureAdapters, "adapters", nil, "Configured adapter names (repeatable or comma-separated)")
	networkConfigureCmd.Flags().BoolVar(&networkConfigureApply, "apply", false, "Apply network configuration changes")
	networkConfigureCmd.Flags().BoolVarP(&networkConfigureForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(networkConfigureCmd, &networkConfigurePlan, &networkConfigureDryRun)
	mergeSchemaAnnotation(networkConfigureCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws", "supervisor"},
		ResourceType: "network",
		InputSources: []string{"flags"},
	})
}

func runNetworkConfigure(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	if len(networkConfigureAdapters) == 0 {
		return fmt.Errorf("at least one adapter is required (--adapters)")
	}

	if planRequested(networkConfigurePlan, networkConfigureDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource": "network",
			},
			Inputs: map[string]any{
				"adapters": networkConfigureAdapters,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send network configuration update with the provided adapter set.",
				"Return updated network configuration metadata.",
			},
			Risks: []string{
				"Network reconfiguration can disrupt connectivity.",
			},
			RequiresConfirmation: !(networkConfigureApply || networkConfigureForce),
			VerificationCommands: []string{
				"hab network get --json",
				"hab system health --json",
			},
		}, textMode, "network")
		return nil
	}

	if !networkConfigureApply {
		return fmt.Errorf("refusing to configure network without --apply")
	}

	if err := confirmAction(networkConfigureForce, "Apply network configuration changes now?", "configure network"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.NetworkConfigure(networkConfigureAdapters)
	if err != nil {
		return err
	}

	output.PrintSuccessWithContext(result, textMode, "Network configuration updated.", output.EnvelopeContext{Operation: "configure", ResourceType: "network"})
	return nil
}

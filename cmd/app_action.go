package cmd

import (
	"fmt"
	"time"

	"github.com/home-assistant/hab/internal/redact"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

type appMutationFlags struct {
	plan, dryRun, force bool
	timeout             time.Duration
}

func (f *appMutationFlags) register(cmd *cobra.Command, destructive bool) {
	registerMutationPlanFlags(cmd, &f.plan, &f.dryRun)
	cmd.Flags().DurationVar(&f.timeout, "timeout", 10*time.Minute, "Maximum time to wait for the app operation")
	sideEffect := "write"
	if destructive {
		sideEffect = "destructive"
		cmd.Flags().BoolVarP(&f.force, "force", "f", false, "Skip confirmation")
	}
	annotateAppCommand(cmd, sideEffect)
}

func newAppActionCommand(operation string) *cobra.Command {
	f := &appMutationFlags{}
	var removeConfig bool
	descriptions := map[string]string{
		"install": "Install an app", "update": "Update an app to its latest version",
		"start": "Start an app", "stop": "Stop an app", "restart": "Restart an app",
		"uninstall": "Uninstall an app",
	}
	cmd := &cobra.Command{
		Use:   operation + " <slug>",
		Short: descriptions[operation],
		Args:  appSlugArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var data map[string]any
			if operation == "uninstall" {
				data = map[string]any{"remove_config": removeConfig}
			}
			return runAppMutation(cmd, args[0], operation, operation, data, f)
		},
	}
	if operation == "uninstall" {
		cmd.Flags().BoolVar(&removeConfig, "remove-config", false, "Also delete the app configuration directory")
	}
	if operation == "update" {
		cmd.Long = "Update an app to its latest version. Create a backup first with 'hab backup create'."
	}
	f.register(cmd, operation == "uninstall" || operation == "stop" || operation == "restart")
	return cmd
}

func runAppMutation(cmd *cobra.Command, slug, operation, endpoint string, data map[string]any, f *appMutationFlags) error {
	if f.timeout <= 0 {
		return fmt.Errorf("--timeout must be positive")
	}
	verification := []string{"hab app get " + slug + " --json"}
	if operation == "uninstall" {
		verification = []string{"hab app list --json"}
	}
	if planRequested(f.plan, f.dryRun) {
		inputs := data
		if operation == "configure" {
			// App schemas can store credentials under arbitrary option names.
			inputs = map[string]any{"configuration": redact.Hidden}
		}
		printMutationPlan(MutationPlan{
			WouldChange:          true,
			Target:               map[string]any{"slug": slug, "operation": operation},
			Inputs:               inputs,
			Steps:                []string{fmt.Sprintf("Ask Supervisor to %s app %s.", operation, slug)},
			RequiresConfirmation: (operation == "uninstall" || operation == "stop" || operation == "restart") && !f.force,
			VerificationCommands: verification,
		}, getTextMode(), "app")
		return nil
	}
	if operation == "uninstall" || operation == "stop" || operation == "restart" {
		if err := confirmAction(f.force, fmt.Sprintf("%s %s?", cmd.Short, slug), operation+" app"); err != nil {
			return err
		}
	}
	if _, err := appRequest(cmd, "post", "/addons/"+slug+"/"+endpoint, data, f.timeout); err != nil {
		return err
	}
	output.PrintSuccessWithContext(nil, getTextMode(), fmt.Sprintf("App '%s': %s completed.", slug, operation), output.EnvelopeContext{
		Operation: operation, ResourceType: "app", VerificationCommands: verification,
	})
	return nil
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	backupCreatePlan   bool
	backupCreateDryRun bool
)

var backupCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new backup",
	Long:  `Create a new backup of Home Assistant.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBackupCreate,
}

func init() {
	backupCmd.AddCommand(backupCreateCmd)
	registerMutationPlanFlags(backupCreateCmd, &backupCreatePlan, &backupCreateDryRun)
	mergeSchemaAnnotation(backupCreateCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "backup",
		InputSources: []string{"args", "flags"},
	})
}

func runBackupCreate(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]interface{}{}
	if len(args) > 0 {
		params["name"] = args[0]
	}

	if planRequested(backupCreatePlan, backupCreateDryRun) {
		target := map[string]any{"resource": "backup"}
		if len(args) > 0 {
			target["name"] = args[0]
		}
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target:      target,
			Inputs:      map[string]any{"params": params},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send backup generate command.",
				"Return backup metadata and status.",
			},
			VerificationCommands: []string{
				"hab backup list --json",
			},
			RequiresConfirmation: false,
		}, textMode, "backup")
		return nil
	}

	result, err := ws.BackupGenerate(params)
	if err != nil {
		return err
	}

	message := "Backup creation initiated."
	if len(args) > 0 {
		message = fmt.Sprintf("Backup '%s' creation initiated.", args[0])
	}
	output.PrintSuccessWithContext(result, textMode, message, output.EnvelopeContext{Operation: "create", ResourceType: "backup"})
	return nil
}

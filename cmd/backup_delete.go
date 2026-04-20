package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	backupDeleteForce  bool
	backupDeletePlan   bool
	backupDeleteDryRun bool
)

var backupDeleteCmd = &cobra.Command{
	Use:   "delete <backup_id>",
	Short: "Delete a backup",
	Long:  `Delete a backup by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupDelete,
}

func init() {
	backupCmd.AddCommand(backupDeleteCmd)
	backupDeleteCmd.Flags().BoolVarP(&backupDeleteForce, "force", "f", false, "Skip confirmation")
	registerMutationPlanFlags(backupDeleteCmd, &backupDeletePlan, &backupDeleteDryRun)
	mergeSchemaAnnotation(backupDeleteCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "backup",
		InputSources: []string{"args", "flags"},
	})
}

func runBackupDelete(cmd *cobra.Command, args []string) error {
	backupID := args[0]
	textMode := getTextMode()

	if planRequested(backupDeletePlan, backupDeleteDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":  "backup",
				"backup_id": backupID,
			},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send backup delete command for the selected backup ID.",
				"Return deletion confirmation.",
			},
			Risks: []string{
				"This operation permanently removes the selected backup.",
			},
			RequiresConfirmation: !backupDeleteForce,
			VerificationCommands: []string{
				"hab backup list --json",
			},
		}, textMode, "backup")
		return nil
	}

	if err := confirmAction(backupDeleteForce, fmt.Sprintf("Delete backup %s?", backupID), "delete backup"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.BackupDelete(backupID); err != nil {
		return err
	}

	output.PrintSuccessWithContext(nil, textMode, fmt.Sprintf("Backup '%s' deleted.", backupID), output.EnvelopeContext{Operation: "delete", ResourceType: "backup"})
	return nil
}

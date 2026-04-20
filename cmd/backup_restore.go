package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	backupRestoreAgent         string
	backupRestorePassword      string
	backupRestoreAddons        []string
	backupRestoreFolders       []string
	backupRestoreDatabase      bool
	backupRestoreHomeAssistant bool
	backupRestoreForce         bool
	backupRestorePlan          bool
	backupRestoreDryRun        bool
)

var backupRestoreCmd = &cobra.Command{
	Use:   "restore <backup_id>",
	Short: "Restore a backup",
	Long:  `Restore a backup by ID using a specific backup agent.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupRestore,
}

func init() {
	backupCmd.AddCommand(backupRestoreCmd)

	backupRestoreCmd.Flags().StringVar(&backupRestoreAgent, "agent", "", "Backup agent ID to restore from")
	backupRestoreCmd.Flags().StringVar(&backupRestorePassword, "password", "", "Backup password")
	backupRestoreCmd.Flags().StringSliceVar(&backupRestoreAddons, "restore-addon", nil, "Addon slug to restore (repeatable)")
	backupRestoreCmd.Flags().StringSliceVar(&backupRestoreFolders, "restore-folder", nil, "Folder to restore (repeatable)")
	backupRestoreCmd.Flags().BoolVar(&backupRestoreDatabase, "restore-database", false, "Restore the Home Assistant database")
	backupRestoreCmd.Flags().BoolVar(&backupRestoreHomeAssistant, "restore-homeassistant", false, "Restore Home Assistant core data")
	backupRestoreCmd.Flags().BoolVarP(&backupRestoreForce, "force", "f", false, "Skip confirmation")
	backupRestoreCmd.MarkFlagRequired("agent")
	registerMutationPlanFlags(backupRestoreCmd, &backupRestorePlan, &backupRestoreDryRun)
	mergeSchemaAnnotation(backupRestoreCmd, SchemaAnnotation{
		SideEffect:   "destructive",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws"},
		ResourceType: "backup",
		InputSources: []string{"args", "flags"},
	})
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
	backupID := args[0]
	textMode := getTextMode()

	params := map[string]interface{}{
		"backup_id": backupID,
		"agent_id":  backupRestoreAgent,
	}

	if backupRestorePassword != "" {
		params["password"] = backupRestorePassword
	}
	if len(backupRestoreAddons) > 0 {
		params["restore_addons"] = backupRestoreAddons
	}
	if len(backupRestoreFolders) > 0 {
		params["restore_folders"] = backupRestoreFolders
	}
	if cmd.Flags().Changed("restore-database") {
		params["restore_database"] = backupRestoreDatabase
	}
	if cmd.Flags().Changed("restore-homeassistant") {
		params["restore_homeassistant"] = backupRestoreHomeAssistant
	}

	if planRequested(backupRestorePlan, backupRestoreDryRun) {
		printMutationPlan(MutationPlan{
			WouldChange: true,
			Target: map[string]any{
				"resource":  "backup",
				"backup_id": backupID,
			},
			Inputs: map[string]any{"params": params},
			Steps: []string{
				"Connect to Home Assistant WebSocket API.",
				"Send backup restore command with selected restore scope.",
				"Return restore job metadata.",
			},
			Risks: []string{
				"Restoring a backup may overwrite current Home Assistant data.",
				"Service interruption is expected during restore.",
			},
			RequiresConfirmation: !backupRestoreForce,
			VerificationCommands: []string{
				"hab system health --json",
				"hab backup list --json",
			},
		}, textMode, "backup")
		return nil
	}

	if err := confirmAction(backupRestoreForce, fmt.Sprintf("Restore backup %s? This can overwrite current data.", backupID), "restore backup"); err != nil {
		return err
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.BackupRestore(params)
	if err != nil {
		return err
	}

	output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("Backup restore started for '%s'.", backupID), output.EnvelopeContext{Operation: "restore", ResourceType: "backup"})
	return nil
}

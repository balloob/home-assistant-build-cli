package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	backupCreateAgents []string
	backupCreatePlan   bool
	backupCreateDryRun bool
)

var backupCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new backup",
	Long: `Create a new backup of Home Assistant.

The backup is stored in the local backup agent unless --agent is given.
Run 'hab backup agents' to list the available agent IDs.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBackupCreate,
}

func init() {
	backupCmd.AddCommand(backupCreateCmd)
	backupCreateCmd.Flags().StringSliceVar(&backupCreateAgents, "agent", nil, "Backup agent ID to store the backup in (repeatable, default: local agent)")
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

	agentIDs := backupCreateAgents
	if len(agentIDs) == 0 {
		localID, err := findLocalBackupAgent(ws)
		if err != nil {
			return err
		}
		agentIDs = []string{localID}
	}

	params := map[string]interface{}{"agent_ids": agentIDs}
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

// findLocalBackupAgent returns the ID of the agent that stores backups on the
// Home Assistant host. Core names it "backup.local"; on Supervisor installs it
// is "hassio.local".
func findLocalBackupAgent(ws interface {
	BackupAgentsInfo() (map[string]interface{}, error)
}) (string, error) {
	info, err := ws.BackupAgentsInfo()
	if err != nil {
		return "", err
	}
	agents, _ := info["agents"].([]interface{})
	found := ""
	for _, a := range agents {
		agent, _ := a.(map[string]interface{})
		id, _ := agent["agent_id"].(string)
		if id == "hassio.local" {
			return id, nil
		}
		if id == "backup.local" {
			found = id
		}
	}
	if found == "" {
		return "", fmt.Errorf("no local backup agent found; pass --agent (see 'hab backup agents')")
	}
	return found, nil
}

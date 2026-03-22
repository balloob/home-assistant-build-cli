package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var backupAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "List backup agents",
	Long:  `List backup agents available for storage and restore operations.`,
	RunE:  runBackupAgents,
}

func init() {
	backupCmd.AddCommand(backupAgentsCmd)
}

func runBackupAgents(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.BackupAgentsInfo()
	if err != nil {
		return err
	}

	if agents, ok := result["agents"].([]interface{}); ok {
		output.PrintOutput(agents, textMode, "")
		return nil
	}

	output.PrintOutput(result, textMode, "")
	return nil
}

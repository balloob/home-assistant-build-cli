package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var repairsListSeverity string

var repairsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List repair issues",
	Long:  `List all repair issues reported by Home Assistant.`,
	Example: `  hab repairs list
  hab repairs list --severity critical
  hab repairs list --severity error
  hab repairs list --json`,
	RunE: runRepairsList,
}

func init() {
	repairsCmd.AddCommand(repairsListCmd)
	repairsListCmd.Flags().StringVar(&repairsListSeverity, "severity", "", "Filter by severity: warning, error, or critical")
}

func runRepairsList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	issues, err := ws.RepairListIssues()
	if err != nil {
		return err
	}

	// Filter by severity if requested
	if repairsListSeverity != "" {
		sev := strings.ToLower(repairsListSeverity)
		var filtered []interface{}
		for _, i := range issues {
			if m, ok := i.(map[string]interface{}); ok {
				if m["severity"] == sev {
					filtered = append(filtered, i)
				}
			}
		}
		issues = filtered
	}

	if len(issues) == 0 {
		output.PrintOutput([]interface{}{}, textMode, "No repair issues found.")
		return nil
	}

	output.PrintOutput(issues, textMode, "")
	return nil
}

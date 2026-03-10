package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var repairsUnignoreCmd = &cobra.Command{
	Use:   "unignore <domain> <issue_id>",
	Short: "Un-ignore a repair issue",
	Long:  `Un-ignore a previously ignored repair issue so it appears as active again.`,
	Example: `  hab repairs unignore homeassistant deprecated_yaml`,
	Args: cobra.ExactArgs(2),
	RunE: runRepairsUnignore,
}

func init() {
	repairsCmd.AddCommand(repairsUnignoreCmd)
}

func runRepairsUnignore(cmd *cobra.Command, args []string) error {
	domain := args[0]
	issueID := args[1]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.RepairIgnoreIssue(domain, issueID, false); err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, "Issue un-ignored.")
	return nil
}

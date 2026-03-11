package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var repairsIgnoreCmd = &cobra.Command{
	Use:   "ignore <domain> <issue_id>",
	Short: "Ignore a repair issue",
	Long:  `Ignore a repair issue so it no longer appears as an active problem.`,
	Example: `  hab repairs ignore homeassistant deprecated_yaml
  hab repairs ignore mqtt broker_connection_failed`,
	Args: cobra.ExactArgs(2),
	RunE: runRepairsIgnore,
}

func init() {
	repairsCmd.AddCommand(repairsIgnoreCmd)
}

func runRepairsIgnore(cmd *cobra.Command, args []string) error {
	domain := args[0]
	issueID := args[1]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ws.RepairIgnoreIssue(domain, issueID, true); err != nil {
		return err
	}

	output.PrintSuccess(nil, textMode, "Issue ignored.")
	return nil
}

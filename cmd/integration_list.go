package cmd

import (
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var integrationListDomain string

var integrationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List integrations",
	Long:  `List all configured integrations (config entries) in Home Assistant.`,
	Example: `  hab integration list
  hab integration list --domain hue
  hab integration list --json`,
	RunE: runIntegrationList,
}

func init() {
	integrationCmd.AddCommand(integrationListCmd)
	integrationListCmd.Flags().StringVar(&integrationListDomain, "domain", "", "Filter by integration domain (e.g. hue, mqtt)")
}

func runIntegrationList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	entries, err := ws.ConfigEntriesList(integrationListDomain)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		output.PrintOutput([]interface{}{}, textMode, "No integrations found.")
		return nil
	}

	output.PrintOutput(entries, textMode, "")
	return nil
}

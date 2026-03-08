package cmd

import (
	"github.com/spf13/cobra"
)

var scriptListFlags *ListFlags

var scriptListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all scripts",
	Long:    `List all scripts in Home Assistant.`,
	GroupID: scriptGroupCommands,
	RunE:    runScriptList,
}

func init() {
	scriptCmd.AddCommand(scriptListCmd)
	scriptListFlags = RegisterListFlags(scriptListCmd, "entity_id")
}

func runScriptList(cmd *cobra.Command, args []string) error {
	return listDomainEntities(stateListConfig{
		domain:       "script",
		listFlags:    scriptListFlags,
		textMode:     getTextMode(),
		emptyMessage: "No scripts.",
	})
}

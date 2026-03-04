package cmd

import (
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

const (
	scriptGroupCommands    = "commands"
	scriptGroupSubcommands = "subcommands"
)

var scriptCmd = &cobra.Command{
	Use:     "script",
	Short:   "Manage scripts",
	Long:    `Create, update, delete, and run scripts.`,
	GroupID: "automation",
}

// resolveScriptConfigID strips the "script." prefix from an entity ID
// to obtain the internal config ID used by the REST API.
func resolveScriptConfigID(_ client.RestAPI, rawID string) (string, error) {
	return strings.TrimPrefix(rawID, "script."), nil
}

func init() {
	rootCmd.AddCommand(scriptCmd)

	scriptCmd.AddGroup(
		&cobra.Group{ID: scriptGroupCommands, Title: "Commands:"},
		&cobra.Group{ID: scriptGroupSubcommands, Title: "Subcommands:"},
	)

	RegisterSubComponentCRUD(SubComponentConfig{
		ParentCmd:       scriptCmd,
		ParentName:      "script",
		ComponentName:   "action",
		ComponentPlural: "actions",
		ConfigKeys:      []string{"sequence"},
		DefaultKey:      "sequence",
		APIBasePath:     "config/script/config/",
		ResolveID:       resolveScriptConfigID,
		ParentFlagName:  "script",
		GroupID:         scriptGroupSubcommands,
	})
}

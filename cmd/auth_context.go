package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/config"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var authContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage Home Assistant contexts",
	Long: `Manage named contexts, one per Home Assistant instance.

Each context stores its own URL and credentials. Log in to a context with
'hab auth login --context <name>', make it the default with
'hab auth context use <name>', or select it for one command with
--context <name> or HAB_CONTEXT=<name>. Remove a context with
'hab auth logout --context <name>'.`,
}

var authContextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List contexts",
	Long:  `List contexts with stored credentials and mark the one selected for this command.`,
	Args:  cobra.NoArgs,
	RunE:  runAuthContextList,
}

var authContextUseCmd = &cobra.Command{
	Use:               "use <name>",
	Short:             "Set the current context",
	Long:              `Set the context used by commands that do not pass --context or HAB_CONTEXT.`,
	Example:           "  hab auth context use brother-homeassistant\n  hab auth context use default",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: contextCompletions,
	RunE:              runAuthContextUse,
}

func init() {
	authCmd.AddCommand(authContextCmd)
	authContextCmd.AddCommand(authContextListCmd)
	authContextCmd.AddCommand(authContextUseCmd)
	mergeSchemaAnnotation(authContextListCmd, SchemaAnnotation{
		SideEffect:   "read",
		Capabilities: []string{"local"},
		ResourceType: "auth",
		GuideTopic:   "auth",
	})
	mergeSchemaAnnotation(authContextUseCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"local"},
		ResourceType: "auth",
		GuideTopic:   "auth",
	})
}

func runAuthContextList(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	names, err := config.ListContexts(configDir)
	if err != nil {
		return err
	}
	selected := getAuthManager().Context
	contexts := make([]map[string]any, 0, len(names))
	for _, name := range names {
		item := map[string]any{"name": name, "current": name == selected}
		// The URL is not secret; tokens are never listed.
		creds, err := auth.LoadStoredCredentials(config.GetContextDir(configDir, name))
		var unreadable *auth.UnreadableCredentialsError
		switch {
		case err == nil && creds != nil:
			item["url"] = creds.URL
		case errors.As(err, &unreadable):
			item["unreadable"] = true
			if unreadable.URL != "" {
				item["url"] = unreadable.URL
			}
		}
		contexts = append(contexts, item)
	}
	output.PrintOutput(contexts, getTextMode(), "")
	return nil
}

func runAuthContextUse(cmd *cobra.Command, args []string) error {
	name := args[0]
	configDir := viper.GetString("config")
	if err := config.ValidateContextName(name); err != nil {
		return err
	}
	names, err := config.ListContexts(configDir)
	if err != nil {
		return err
	}
	if name != config.DefaultContext && !slices.Contains(names, name) {
		available := "none"
		if len(names) > 0 {
			available = strings.Join(names, ", ")
		}
		return fmt.Errorf("%w: context %q has no stored credentials (available contexts: %s). Run 'hab auth login --context %s' first", auth.ErrNotAuthenticated, name, available, name)
	}
	if err := config.SaveCurrentContext(configDir, name); err != nil {
		return err
	}
	output.PrintSuccess(map[string]any{"context": name}, getTextMode(), fmt.Sprintf("Current context set to %s.", name))
	return nil
}

func contextCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names, _ := config.ListContexts(viper.GetString("config"))
	return names, cobra.ShellCompDirectiveNoFileComp
}

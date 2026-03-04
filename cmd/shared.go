package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getWSClient creates an authenticated, connected WebSocket client.
// Caller must defer ws.Close() after a successful return.
func getWSClient() (client.WebSocketAPI, error) {
	configDir := viper.GetString("config")
	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return nil, err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return nil, err
	}
	return ws, nil
}

// getRESTClient creates an authenticated REST client.
func getRESTClient() (client.RestAPI, error) {
	configDir := viper.GetString("config")
	manager := auth.NewManager(configDir)
	return manager.GetRestClient()
}

// getCredentials returns the current authentication credentials.
func getCredentials() (*auth.Credentials, error) {
	configDir := viper.GetString("config")
	manager := auth.NewManager(configDir)
	return manager.GetCredentials()
}

// getTextMode returns whether text output mode is enabled.
func getTextMode() bool {
	return viper.GetBool("text")
}

// resolveArg resolves a value from either a flag variable or a positional argument.
// It checks the flag value first; if empty, falls back to args[index].
// Returns an error if no value is found.
func resolveArg(flagVal string, args []string, index int, name string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	if len(args) > index {
		return args[index], nil
	}
	return "", fmt.Errorf("%s is required", name)
}

// confirmAction prompts the user for confirmation unless force or textMode is set.
// Returns true if the action should proceed, false if the user declined.
func confirmAction(force, textMode bool, prompt string) bool {
	if force || textMode {
		return true
	}
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// ---------------------------------------------------------------------------
// List output helpers
// ---------------------------------------------------------------------------

// ListFlags holds the common --count, --brief, --limit flag values
// used by list commands.
type ListFlags struct {
	Count bool
	Brief bool
	Limit int
}

// RegisterListFlags adds --count/-c, --brief/-b, --limit/-n flags to a command
// and returns a ListFlags struct that will be populated when the command runs.
func RegisterListFlags(cmd *cobra.Command, idField string) *ListFlags {
	f := &ListFlags{}
	cmd.Flags().BoolVarP(&f.Count, "count", "c", false, "Return only the count of items")
	cmd.Flags().BoolVarP(&f.Brief, "brief", "b", false,
		fmt.Sprintf("Return minimal fields (%s and name only)", idField))
	cmd.Flags().IntVarP(&f.Limit, "limit", "n", 0, "Limit results to N items")
	return f
}

// RenderCount outputs the item count and returns true if the Count flag is set.
func (f *ListFlags) RenderCount(count int, textMode bool) bool {
	if !f.Count {
		return false
	}
	if textMode {
		fmt.Printf("Count: %d\n", count)
	} else {
		output.PrintOutput(map[string]interface{}{"count": count}, false, "")
	}
	return true
}

// ApplyLimit truncates items to the Limit if set ([]interface{} variant).
func (f *ListFlags) ApplyLimit(items []interface{}) []interface{} {
	if f.Limit > 0 && len(items) > f.Limit {
		return items[:f.Limit]
	}
	return items
}

// ApplyLimitMap truncates items to the Limit if set ([]map variant).
func (f *ListFlags) ApplyLimitMap(items []map[string]interface{}) []map[string]interface{} {
	if f.Limit > 0 && len(items) > f.Limit {
		return items[:f.Limit]
	}
	return items
}

// RenderBrief outputs brief items and returns true if the Brief flag is set.
// Text mode prints "name (id)" per line; JSON mode outputs [{idField, nameField}].
// Works with []interface{} where each element is map[string]interface{}.
func (f *ListFlags) RenderBrief(items []interface{}, textMode bool, idField, nameField string) bool {
	if !f.Brief {
		return false
	}
	if textMode {
		for _, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				name, _ := m[nameField].(string)
				id, _ := m[idField].(string)
				fmt.Printf("%s (%s)\n", name, id)
			}
		}
	} else {
		var brief []map[string]interface{}
		for _, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				brief = append(brief, map[string]interface{}{
					idField:   m[idField],
					nameField: m[nameField],
				})
			}
		}
		output.PrintOutput(brief, false, "")
	}
	return true
}

// RenderBriefMap is the same as RenderBrief but for []map[string]interface{}.
func (f *ListFlags) RenderBriefMap(items []map[string]interface{}, textMode bool, idField, nameField string) bool {
	if !f.Brief {
		return false
	}
	if textMode {
		for _, item := range items {
			name, _ := item[nameField].(string)
			id, _ := item[idField].(string)
			fmt.Printf("%s (%s)\n", name, id)
		}
	} else {
		var brief []map[string]interface{}
		for _, item := range items {
			brief = append(brief, map[string]interface{}{
				idField:   item[idField],
				nameField: item[nameField],
			})
		}
		output.PrintOutput(brief, false, "")
	}
	return true
}

// ---------------------------------------------------------------------------
// Input helpers
// ---------------------------------------------------------------------------

// InputFlags holds the standard --data, --file, --format flag values
// used by commands that accept JSON/YAML configuration input.
type InputFlags struct {
	Data   string
	File   string
	Format string
}

// Register adds the --data/-d, --file/-f, and --format flags to a cobra command.
func (f *InputFlags) Register(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.Data, "data", "d", "", "Configuration data as JSON or YAML string")
	cmd.Flags().StringVarP(&f.File, "file", "f", "", "Path to configuration file")
	cmd.Flags().StringVar(&f.Format, "format", "", "Input format: json or yaml (auto-detected if not specified)")
}

// Parse parses the input data from the flag values, returning a map.
func (f *InputFlags) Parse() (map[string]interface{}, error) {
	return input.ParseInput(f.Data, f.File, f.Format)
}

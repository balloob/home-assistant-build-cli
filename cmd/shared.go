package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Shared cobra group IDs for parent commands that split their subcommands
// into "commands" (top-level actions) and "subcommands" (resource-specific
// CRUD). Each parent (automation, script, dashboard, helper) uses these
// values to group its children consistently in --help output.
const (
	groupCommands    = "commands"
	groupSubcommands = "subcommands"
)

// helperDomains is the set of storage-based helper entity domains.
// Used by helper list, overview, and other commands that need to identify
// helper entities by their domain prefix.
var helperDomains = map[string]bool{
	"input_boolean":  true,
	"input_number":   true,
	"input_text":     true,
	"input_select":   true,
	"input_datetime": true,
	"input_button":   true,
	"counter":        true,
	"timer":          true,
	"schedule":       true,
}

// authManagerOnce ensures the auth.Manager is created once per CLI invocation,
// avoiding redundant credential file reads and decryption when commands use
// both REST and WebSocket clients.
var (
	authManagerOnce sync.Once
	cachedAuthMgr   *auth.Manager
)

// getAuthManager returns a cached auth.Manager using the configured config dir.
func getAuthManager() *auth.Manager {
	authManagerOnce.Do(func() {
		cachedAuthMgr = auth.NewManager(viper.GetString("config"))
	})
	return cachedAuthMgr
}

// getWSClient creates an authenticated, connected WebSocket client.
// Caller must defer ws.Close() after a successful return.
func getWSClient() (client.WebSocketAPI, error) {
	manager := getAuthManager()
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
	return getAuthManager().GetRestClient()
}

// getCredentials returns the current authentication credentials.
func getCredentials() (*auth.Credentials, error) {
	return getAuthManager().GetCredentials()
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

// ensureDomainPrefix ensures an entity ID has the given domain prefix.
// If id already starts with "domain.", it is returned unchanged.
func ensureDomainPrefix(id, domain string) string {
	prefix := domain + "."
	if strings.HasPrefix(id, prefix) {
		return id
	}
	return prefix + id
}

// ---------------------------------------------------------------------------
// Helper delete orchestration
// ---------------------------------------------------------------------------

// deleteHelperByEntityOrEntryID deletes a helper by entity_id, helper ID, or
// config_entry_id. Storage-based helpers (input_boolean, counter, …) are
// deleted via the WS helper/delete command; config-entry-based helpers (group)
// are deleted via the config entry API. This orchestration logic was moved
// from the client package to keep domain policy in cmd/.
func deleteHelperByEntityOrEntryID(ws client.WebSocketAPI, id, helperType string) error {
	isEntityID := strings.Contains(id, ".")

	// Storage-based helpers: use HelperDelete
	if helperDomains[helperType] {
		helperID := id
		if isEntityID {
			if _, after, ok := strings.Cut(id, "."); ok {
				helperID = after
			}
		}
		return ws.HelperDelete(helperType, helperID)
	}

	// Config-entry-based helpers (e.g. group): resolve to config_entry_id
	var entryID string
	if isEntityID {
		resolved, err := ws.ResolveEntityToConfigEntry(id)
		if err != nil {
			return fmt.Errorf("failed to resolve entity_id: %w", err)
		}
		if resolved == "" {
			return fmt.Errorf("entity %s does not have a config entry", id)
		}
		entryID = resolved
	} else {
		entryID = id
	}

	return ws.ConfigEntryDelete(entryID)
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

// applyLimit is the generic implementation shared by both slice types.
func applyLimit[T any](items []T, limit int) []T {
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

// ApplyLimit truncates items to the Limit if set ([]interface{} variant).
func (f *ListFlags) ApplyLimit(items []interface{}) []interface{} {
	return applyLimit(items, f.Limit)
}

// ApplyLimitMap truncates items to the Limit if set ([]map variant).
func (f *ListFlags) ApplyLimitMap(items []map[string]interface{}) []map[string]interface{} {
	return applyLimit(items, f.Limit)
}

// renderBriefCore implements the shared brief-rendering logic.
// extractFields returns (id, name) from each item.
func renderBriefCore[T any](items []T, textMode bool, idField, nameField string, extractFields func(T) (id, name string), buildBrief func(T) map[string]interface{}) {
	if textMode {
		for _, item := range items {
			name, id := extractFields(item)
			fmt.Printf("%s (%s)\n", name, id)
		}
	} else {
		brief := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			brief = append(brief, buildBrief(item))
		}
		output.PrintOutput(brief, false, "")
	}
}

// RenderBrief outputs brief items and returns true if the Brief flag is set.
// Text mode prints "name (id)" per line; JSON mode outputs [{idField, nameField}].
// Works with []interface{} where each element is map[string]interface{}.
func (f *ListFlags) RenderBrief(items []interface{}, textMode bool, idField, nameField string) bool {
	if !f.Brief {
		return false
	}
	renderBriefCore(items, textMode, idField, nameField,
		func(item interface{}) (string, string) {
			if m, ok := item.(map[string]interface{}); ok {
				name, _ := m[nameField].(string)
				id, _ := m[idField].(string)
				return name, id
			}
			return "", ""
		},
		func(item interface{}) map[string]interface{} {
			if m, ok := item.(map[string]interface{}); ok {
				return map[string]interface{}{
					idField:   m[idField],
					nameField: m[nameField],
				}
			}
			return nil
		},
	)
	return true
}

// RenderBriefMap is the same as RenderBrief but for []map[string]interface{}.
func (f *ListFlags) RenderBriefMap(items []map[string]interface{}, textMode bool, idField, nameField string) bool {
	if !f.Brief {
		return false
	}
	renderBriefCore(items, textMode, idField, nameField,
		func(item map[string]interface{}) (string, string) {
			name, _ := item[nameField].(string)
			id, _ := item[idField].(string)
			return name, id
		},
		func(item map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				idField:   item[idField],
				nameField: item[nameField],
			}
		},
	)
	return true
}

// RenderBriefFields outputs brief items and returns true if the Brief flag is set.
// Text mode prints "name (id)" per line; JSON mode outputs items with only
// the specified jsonFields. This variant is useful when brief JSON output
// should include a configurable set of fields beyond just id and name.
func (f *ListFlags) RenderBriefFields(items []interface{}, textMode bool, idField, nameField string, jsonFields []string) bool {
	if !f.Brief {
		return false
	}
	renderBriefCore(items, textMode, idField, nameField,
		func(item interface{}) (string, string) {
			if m, ok := item.(map[string]interface{}); ok {
				name, _ := m[nameField].(string)
				id, _ := m[idField].(string)
				return name, id
			}
			return "", ""
		},
		func(item interface{}) map[string]interface{} {
			if m, ok := item.(map[string]interface{}); ok {
				b := make(map[string]interface{})
				for _, field := range jsonFields {
					b[field] = m[field]
				}
				return b
			}
			return nil
		},
	)
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

package cmd

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
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

	stdinIsTerminalFunc  = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
	stdoutIsTerminalFunc = func() bool { return term.IsTerminal(int(os.Stdout.Fd())) }
	readConfirmationLine = func() (string, error) {
		reader := bufio.NewReader(os.Stdin)
		return reader.ReadString('\n')
	}

	executionMetadataMu sync.Mutex
	executionMetadata   = map[string]any{}
)

func init() {
	output.SetDefaultMetadataProvider(getExecutionMetadata)
	output.SetDefaultEnvelopeContextProvider(getExecutionEnvelopeContext)
}

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
	noteAuthSource(creds.Source)
	noteTransport("websocket")

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return nil, err
	}
	return ws, nil
}

// getRESTClient creates an authenticated REST client.
func getRESTClient() (client.RestAPI, error) {
	restClient, err := getAuthManager().GetRestClient()
	if err != nil {
		return nil, err
	}
	if creds, credsErr := getAuthManager().GetCredentials(); credsErr == nil && creds != nil {
		noteAuthSource(creds.Source)
	}
	noteTransport("rest")
	return restClient, nil
}

// getCredentials returns the current authentication credentials.
func getCredentials() (*auth.Credentials, error) {
	creds, err := getAuthManager().GetCredentials()
	if err == nil && creds != nil {
		noteAuthSource(creds.Source)
	}
	return creds, err
}

// getTextMode returns whether text output mode is enabled.
func getTextMode() bool {
	return viper.GetBool("text")
}

func isInteractiveInput() bool {
	return stdinIsTerminalFunc()
}

func isInteractiveOutput() bool {
	return stdoutIsTerminalFunc()
}

func resetExecutionMetadata() {
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	executionMetadata = map[string]any{}
}

func setExecutionMetadata(key string, value any) {
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	executionMetadata[key] = value
}

func noteTransport(name string) {
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	transports, _ := executionMetadata["transports_used"].([]string)
	for _, existing := range transports {
		if existing == name {
			return
		}
	}
	executionMetadata["transports_used"] = append(transports, name)
}

func noteAuthSource(source string) {
	if source == "" {
		return
	}
	setExecutionMetadata("auth_source", source)
}

func noteResolution(key, value string) {
	if key == "" || value == "" {
		return
	}
	normalizedKey := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, " ", "_"), "-", "_"))
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	resolutions, _ := executionMetadata["resolved_inputs"].(map[string]any)
	if resolutions == nil {
		resolutions = map[string]any{}
	}
	resolutions[normalizedKey] = value
	executionMetadata["resolved_inputs"] = resolutions
}

func noteFallback(message string) {
	if message == "" {
		return
	}
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	fallbacks, _ := executionMetadata["fallbacks_applied"].([]string)
	executionMetadata["fallbacks_applied"] = append(fallbacks, message)
	executionMetadata["partial_result"] = true
	warnings, _ := executionMetadata["warnings"].([]string)
	executionMetadata["warnings"] = append(warnings, message)
}

func noteWarning(message string) {
	if message == "" {
		return
	}
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	warnings, _ := executionMetadata["warnings"].([]string)
	executionMetadata["warnings"] = append(warnings, message)
}

func noteMissingSection(section string) {
	if section == "" {
		return
	}
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	missing, _ := executionMetadata["missing_sections"].([]string)
	executionMetadata["missing_sections"] = append(missing, section)
	executionMetadata["partial_result"] = true
}

func noteOutputMode(mode string) {
	setExecutionMetadata("output_mode", mode)
	setExecutionMetadata("interactive_input", isInteractiveInput())
	setExecutionMetadata("interactive_output", isInteractiveOutput())
}

func getExecutionMetadata() map[string]interface{} {
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	if len(executionMetadata) == 0 {
		return nil
	}
	return maps.Clone(executionMetadata)
}

func getExecutionEnvelopeContext() output.EnvelopeContext {
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	ctx := output.EnvelopeContext{}
	if op, ok := executionMetadata["default_operation"].(string); ok {
		ctx.Operation = op
	}
	if resourceType, ok := executionMetadata["default_resource_type"].(string); ok {
		ctx.ResourceType = resourceType
	}
	if partial, ok := executionMetadata["partial_result"].(bool); ok {
		ctx.PartialResult = partial
	}
	if warnings, ok := executionMetadata["warnings"].([]string); ok {
		ctx.Warnings = append([]string(nil), warnings...)
	}
	if fallbacks, ok := executionMetadata["fallbacks_applied"].([]string); ok {
		ctx.FallbacksApplied = append([]string(nil), fallbacks...)
	}
	if missing, ok := executionMetadata["missing_sections"].([]string); ok {
		ctx.MissingSections = append([]string(nil), missing...)
	}
	return ctx
}

func noteDefaultEnvelope(operation, resourceType string) {
	if operation != "" {
		setExecutionMetadata("default_operation", operation)
	}
	if resourceType != "" {
		setExecutionMetadata("default_resource_type", resourceType)
	}
}

func determineOutputMode(cmd *cobra.Command) (text bool, json bool, mode string, err error) {
	jsonFlag := cmd.Flags().Lookup("json")
	textFlag := cmd.Flags().Lookup("text")
	jsonExplicit := jsonFlag != nil && jsonFlag.Changed
	textExplicit := textFlag != nil && textFlag.Changed
	jsonRequested := viper.GetBool("json")
	textRequested := viper.GetBool("text")

	if jsonExplicit && textExplicit && jsonRequested == textRequested {
		return false, false, "", fmt.Errorf("conflicting output flags: use either --json or --text")
	}

	switch {
	case jsonExplicit && jsonRequested:
		return false, true, "json", nil
	case textExplicit && textRequested:
		return true, false, "text", nil
	case !isInteractiveOutput():
		return false, true, "json", nil
	default:
		return true, false, "text", nil
	}
}

// resolveArg resolves a value from either a flag variable or a positional argument.
// If both sources are present, they must match exactly.
func resolveArg(flagVal string, args []string, index int, name string) (string, error) {
	if flagVal != "" && len(args) > index {
		if args[index] != flagVal {
			return "", fmt.Errorf("conflicting %s values: positional %q does not match flag value %q", name, args[index], flagVal)
		}
		noteResolution(name, "positional_and_flag")
		return flagVal, nil
	}
	if flagVal != "" {
		noteResolution(name, "flag")
		return flagVal, nil
	}
	if len(args) > index {
		noteResolution(name, "positional")
		return args[index], nil
	}
	return "", fmt.Errorf("%s is required; provide it as a positional argument or matching flag value", name)
}

// confirmAction prompts the user for confirmation unless force is set.
// In non-interactive mode it returns a structured confirmation-required error.
func confirmAction(force bool, prompt, action string) error {
	if force {
		return nil
	}
	if !isInteractiveInput() {
		return client.NewConfirmationRequiredError(action, prompt)
	}
	fmt.Printf("%s [y/N]: ", prompt)
	response, err := readConfirmationLine()
	if err != nil {
		return client.NewConfirmationRequiredError(action, prompt)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		return nil
	}
	return client.NewCancelledError(action)
}

func cancelledError(action string) error {
	return client.NewCancelledError(action)
}

// MutationPlan describes a planned mutation without applying it.
type MutationPlan struct {
	Mode                 string         `json:"mode"`
	WouldChange          bool           `json:"would_change"`
	Target               map[string]any `json:"target"`
	Inputs               map[string]any `json:"inputs,omitempty"`
	DerivedIDs           map[string]any `json:"derived_ids,omitempty"`
	Steps                []string       `json:"steps"`
	Risks                []string       `json:"risks,omitempty"`
	RequiresConfirmation bool           `json:"requires_confirmation"`
	VerificationCommands []string       `json:"verification_commands,omitempty"`
}

func registerMutationPlanFlags(cmd *cobra.Command, planFlag, dryRunFlag *bool) {
	cmd.Flags().BoolVar(planFlag, "plan", false, "Show execution plan without applying changes")
	cmd.Flags().BoolVar(dryRunFlag, "dry-run", false, "Alias for --plan")
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		InputSources:   []string{"flags"},
		OutputVariants: []string{"plan", "full"},
		FlagConstraints: []SchemaFlagConstraint{{
			Type:        "equivalent",
			Flags:       []string{"plan", "dry-run"},
			Description: "--dry-run is an alias of --plan.",
		}},
	})
}

func planRequested(planFlag, dryRunFlag bool) bool {
	return planFlag || dryRunFlag
}

func printMutationPlan(plan MutationPlan, textMode bool, resourceType string) {
	if plan.Mode == "" {
		plan.Mode = "plan"
	}
	output.PrintOutputWithContext(plan, textMode, "Plan only. No changes were applied.", output.EnvelopeContext{
		Operation:    "plan",
		ResourceType: resourceType,
	})
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
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		OutputVariants: []string{"count", "brief", "full"},
		InputSources:   []string{"flags"},
	})
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
// Service call helpers
// ---------------------------------------------------------------------------

// callServiceAction calls a Home Assistant service action via REST and prints
// a success message. It is a convenience wrapper used by commands that map
// directly to a single service call (todo, notification, calendar, etc.).
func callServiceAction(domain, service, successMsg string, data map[string]interface{}) error {
	textMode := getTextMode()
	restClient, err := getRESTClient()
	if err != nil {
		return err
	}
	if _, err := restClient.CallService(domain, service, data); err != nil {
		return err
	}
	output.PrintSuccessWithContext(nil, textMode, successMsg, output.EnvelopeContext{
		Operation:    "call_service",
		ResourceType: domain + "." + service,
	})
	return nil
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
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		InputSources: []string{"flags", "data", "file"},
		FlagConstraints: []SchemaFlagConstraint{{
			Type:        "one_of",
			Flags:       []string{"data", "file"},
			Description: "Use inline data or a file path as the payload source.",
		}},
	})
}

// Parse parses the input data from the flag values, returning a map.
func (f *InputFlags) Parse() (map[string]interface{}, error) {
	return input.ParseInput(f.Data, f.File, f.Format)
}

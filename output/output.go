// Package output formats command results as JSON response envelopes or
// human-readable text for terminal display.
package output

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	defaultMetadataMu       sync.RWMutex
	defaultMetadataProvider func() map[string]interface{}
	defaultContextMu        sync.RWMutex
	defaultContextProvider  func() EnvelopeContext
)

// Text-mode table formatting limits. These keep terminal output readable
// without overwhelming the screen.
const (
	// maxTableColumns is the maximum number of columns shown in a table.
	maxTableColumns = 6
	// maxTableRows is the maximum number of rows shown before truncation.
	maxTableRows = 50
	// maxCellWidth is the maximum character width of a single table cell
	// before the value is truncated with "...".
	maxCellWidth = 30
	// maxDictArrayItems is the maximum number of array items shown when
	// rendering a dict value that contains an array.
	maxDictArrayItems = 10
)

// Response represents the standard JSON output format
type Response struct {
	Success               bool                   `json:"success"`
	Operation             string                 `json:"operation"`
	ResourceType          string                 `json:"resource_type"`
	Data                  interface{}            `json:"data"`
	Message               string                 `json:"message"`
	PartialResult         bool                   `json:"partial_result"`
	Warnings              []string               `json:"warnings"`
	FallbacksApplied      []string               `json:"fallbacks_applied"`
	MissingSections       []string               `json:"missing_sections"`
	VerificationCommands  []string               `json:"verification_commands"`
	NextSuggestedCommands []string               `json:"next_suggested_commands"`
	Error                 *ErrorDetail           `json:"error"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// EnvelopeContext contains optional machine-readable metadata for JSON responses.
// These fields are ignored in text mode output.
type EnvelopeContext struct {
	Operation             string
	ResourceType          string
	PartialResult         bool
	Warnings              []string
	FallbacksApplied      []string
	MissingSections       []string
	VerificationCommands  []string
	NextSuggestedCommands []string
	Metadata              map[string]interface{}
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// FormatOutput formats data for output
func FormatOutput(data interface{}, textMode bool, message string) string {
	return FormatOutputWithContext(data, textMode, message, EnvelopeContext{})
}

// FormatOutputWithContext formats data for output with optional JSON metadata.
func FormatOutputWithContext(data interface{}, textMode bool, message string, ctx EnvelopeContext) string {
	if textMode {
		return formatText(data, message)
	}
	return formatJSON(data, true, message, nil, ctx)
}

// FormatSuccess formats a successful response
func FormatSuccess(data interface{}, message string) string {
	return FormatSuccessWithContext(data, message, EnvelopeContext{})
}

// FormatSuccessWithContext formats a successful JSON response with optional metadata.
func FormatSuccessWithContext(data interface{}, message string, ctx EnvelopeContext) string {
	return formatJSON(data, true, message, nil, ctx)
}

// FormatError formats an error response
func FormatError(code string, msg string, details map[string]interface{}) string {
	return FormatErrorWithContext(code, msg, details, EnvelopeContext{})
}

// FormatErrorWithContext formats an error JSON response with optional metadata.
func FormatErrorWithContext(code string, msg string, details map[string]interface{}, ctx EnvelopeContext) string {
	return formatJSON(nil, false, "", &ErrorDetail{
		Code:    code,
		Message: msg,
		Details: details,
	}, ctx)
}

// FormatErrorText formats an error for text output
func FormatErrorText(msg string, suggestion string) string {
	output := fmt.Sprintf("Error: %s", msg)
	if suggestion != "" {
		output += fmt.Sprintf("\nSuggestion: %s", suggestion)
	}
	return output
}

func formatJSON(data interface{}, success bool, message string, errDetail *ErrorDetail, ctx EnvelopeContext) string {
	defaultContextMu.RLock()
	contextProvider := defaultContextProvider
	defaultContextMu.RUnlock()
	if contextProvider != nil {
		ctx = mergeEnvelopeContext(contextProvider(), ctx)
	}

	if len(ctx.NextSuggestedCommands) == 0 && len(ctx.VerificationCommands) > 0 {
		ctx.NextSuggestedCommands = append([]string(nil), ctx.VerificationCommands...)
	}

	metadata := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	defaultMetadataMu.RLock()
	provider := defaultMetadataProvider
	defaultMetadataMu.RUnlock()
	if provider != nil {
		for key, value := range provider() {
			metadata[key] = value
		}
	}
	if len(ctx.Metadata) > 0 {
		for key, value := range maps.Clone(ctx.Metadata) {
			metadata[key] = value
		}
	}

	resp := Response{
		Success:               success,
		Operation:             ctx.Operation,
		ResourceType:          ctx.ResourceType,
		Data:                  data,
		Message:               message,
		PartialResult:         ctx.PartialResult,
		Warnings:              normalizedStrings(ctx.Warnings),
		FallbacksApplied:      normalizedStrings(ctx.FallbacksApplied),
		MissingSections:       normalizedStrings(ctx.MissingSections),
		VerificationCommands:  normalizedStrings(ctx.VerificationCommands),
		NextSuggestedCommands: normalizedStrings(ctx.NextSuggestedCommands),
		Error:                 errDetail,
		Metadata:              metadata,
	}

	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"success": false, "error": {"code": "MARSHAL_ERROR", "message": %q}}`, err.Error())
	}
	return string(b)
}

func normalizedStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// SetDefaultMetadataProvider configures process-wide JSON metadata that will be
// merged into every response envelope before command-specific metadata.
func SetDefaultMetadataProvider(provider func() map[string]interface{}) {
	defaultMetadataMu.Lock()
	defer defaultMetadataMu.Unlock()
	defaultMetadataProvider = provider
}

// SetDefaultEnvelopeContextProvider configures process-wide JSON envelope fields
// that will be merged into every response envelope before command-specific values.
func SetDefaultEnvelopeContextProvider(provider func() EnvelopeContext) {
	defaultContextMu.Lock()
	defer defaultContextMu.Unlock()
	defaultContextProvider = provider
}

func mergeEnvelopeContext(base, override EnvelopeContext) EnvelopeContext {
	merged := base
	if override.Operation != "" {
		merged.Operation = override.Operation
	}
	if override.ResourceType != "" {
		merged.ResourceType = override.ResourceType
	}
	merged.PartialResult = merged.PartialResult || override.PartialResult
	merged.Warnings = appendUniqueStrings(merged.Warnings, override.Warnings)
	merged.FallbacksApplied = appendUniqueStrings(merged.FallbacksApplied, override.FallbacksApplied)
	merged.MissingSections = appendUniqueStrings(merged.MissingSections, override.MissingSections)
	merged.VerificationCommands = appendUniqueStrings(merged.VerificationCommands, override.VerificationCommands)
	merged.NextSuggestedCommands = appendUniqueStrings(merged.NextSuggestedCommands, override.NextSuggestedCommands)
	if len(base.Metadata) > 0 {
		merged.Metadata = maps.Clone(base.Metadata)
	}
	if len(override.Metadata) > 0 {
		if merged.Metadata == nil {
			merged.Metadata = map[string]interface{}{}
		}
		for key, value := range override.Metadata {
			merged.Metadata[key] = value
		}
	}
	return merged
}

func appendUniqueStrings(existing, incoming []string) []string {
	if len(incoming) == 0 {
		return existing
	}
	seen := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		seen[item] = struct{}{}
	}
	for _, item := range incoming {
		if _, ok := seen[item]; ok {
			continue
		}
		existing = append(existing, item)
		seen[item] = struct{}{}
	}
	return existing
}

func formatText(data interface{}, message string) string {
	if message != "" {
		return message
	}

	if data == nil {
		return "Done."
	}

	switch v := data.(type) {
	case string:
		return v
	case bool:
		if v {
			return "Yes"
		}
		return "No"
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case []interface{}:
		return formatList(v)
	case map[string]interface{}:
		return formatDict(v)
	default:
		// Try to marshal and re-parse as generic type
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		var parsed interface{}
		if err := json.Unmarshal(b, &parsed); err != nil {
			return string(b)
		}
		return formatText(parsed, "")
	}
}

func formatList(data []interface{}) string {
	if len(data) == 0 {
		return "No items."
	}

	// Check if list of maps with common keys
	allMaps := true
	for _, item := range data {
		if _, ok := item.(map[string]interface{}); !ok {
			allMaps = false
			break
		}
	}

	if allMaps {
		return formatDictList(data)
	}

	// Simple list
	var lines []string
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			name := getDisplayName(m)
			lines = append(lines, fmt.Sprintf("  - %s", name))
		} else {
			lines = append(lines, fmt.Sprintf("  - %v", item))
		}
	}
	return strings.Join(lines, "\n")
}

func formatDictList(data []interface{}) string {
	if len(data) == 0 {
		return "No items."
	}

	// Get keys from first item
	first, ok := data[0].(map[string]interface{})
	if !ok {
		return "No items."
	}

	var keys []string
	for k := range first {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) > maxTableColumns {
		keys = keys[:maxTableColumns]
	}

	var lines []string

	// Header
	var headerParts []string
	for _, k := range keys {
		headerParts = append(headerParts, formatKey(k))
	}
	header := strings.Join(headerParts, " | ")
	lines = append(lines, header)
	lines = append(lines, strings.Repeat("-", len(header)))

	// Rows
	for i, item := range data {
		if i >= maxTableRows {
			break
		}
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var values []string
		for _, k := range keys {
			v := m[k]
			str := formatValue(v)
			if len(str) > maxCellWidth {
				str = str[:maxCellWidth-3] + "..."
			}
			values = append(values, str)
		}
		lines = append(lines, strings.Join(values, " | "))
	}

	if len(data) > maxTableRows {
		lines = append(lines, fmt.Sprintf("... and %d more items", len(data)-maxTableRows))
	}

	return strings.Join(lines, "\n")
}

func formatDict(data map[string]interface{}) string {
	var lines []string

	// Sort keys for deterministic output
	sortedKeys := make([]string, 0, len(data))
	for k := range data {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		value := data[key]
		keyLabel := formatKey(key)

		switch v := value.(type) {
		case map[string]interface{}:
			lines = append(lines, fmt.Sprintf("%s:", keyLabel))
			innerKeys := make([]string, 0, len(v))
			for k := range v {
				innerKeys = append(innerKeys, k)
			}
			sort.Strings(innerKeys)
			for _, k := range innerKeys {
				lines = append(lines, fmt.Sprintf("  %s: %s", k, formatValue(v[k])))
			}
		case []interface{}:
			lines = append(lines, fmt.Sprintf("%s:", keyLabel))
			for i, item := range v {
				if i >= maxDictArrayItems {
					lines = append(lines, fmt.Sprintf("  ... and %d more", len(v)-maxDictArrayItems))
					break
				}
				if m, ok := item.(map[string]interface{}); ok {
					lines = append(lines, fmt.Sprintf("  - %s", getDisplayName(m)))
				} else {
					lines = append(lines, fmt.Sprintf("  - %v", item))
				}
			}
		default:
			lines = append(lines, fmt.Sprintf("%s: %s", keyLabel, formatValue(value)))
		}
	}

	return strings.Join(lines, "\n")
}

func formatKey(key string) string {
	// Convert snake_case to Title Case
	parts := strings.Split(key, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case bool:
		if val {
			return "Yes"
		}
		return "No"
	case map[string]interface{}, []interface{}:
		return "..."
	default:
		return fmt.Sprintf("%v", val)
	}
}

func getDisplayName(m map[string]interface{}) string {
	// Try common name fields
	for _, key := range []string{"friendly_name", "name", "entity_id", "id"} {
		if v, ok := m[key]; ok {
			return fmt.Sprintf("%v", v)
		}
	}
	return fmt.Sprintf("%v", m)
}

// PrintOutput prints formatted output to stdout.
// In text mode, it formats data as human-readable text.
// In JSON mode, it wraps data in a success response envelope.
func PrintOutput(data interface{}, textMode bool, message string) {
	PrintOutputWithContext(data, textMode, message, EnvelopeContext{})
}

// PrintOutputWithContext prints formatted output to stdout with optional JSON metadata.
func PrintOutputWithContext(data interface{}, textMode bool, message string, ctx EnvelopeContext) {
	output := FormatOutputWithContext(data, textMode, message, ctx)
	fmt.Println(output)
}

// PrintSuccess prints a successful response with a message.
// Behaves identically to PrintOutput — kept for semantic clarity at call sites.
func PrintSuccess(data interface{}, textMode bool, message string) {
	PrintSuccessWithContext(data, textMode, message, EnvelopeContext{})
}

// PrintSuccessWithContext prints a successful response with optional JSON metadata.
func PrintSuccessWithContext(data interface{}, textMode bool, message string, ctx EnvelopeContext) {
	if textMode {
		fmt.Println(formatText(data, message))
	} else {
		fmt.Println(FormatSuccessWithContext(data, message, ctx))
	}
}

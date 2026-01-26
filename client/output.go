package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Response represents the standard JSON output format
type Response struct {
	Success  bool                   `json:"success"`
	Data     interface{}            `json:"data,omitempty"`
	Message  string                 `json:"message,omitempty"`
	Error    *ErrorDetail           `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// FormatOutput formats data for output
func FormatOutput(data interface{}, textMode bool, message string) string {
	if textMode {
		return formatText(data, message)
	}
	return formatJSON(data, true, message, nil)
}

// FormatSuccess formats a successful response
func FormatSuccess(data interface{}, message string) string {
	return formatJSON(data, true, message, nil)
}

// FormatError formats an error response
func FormatError(code string, msg string, details map[string]interface{}) string {
	return formatJSON(nil, false, "", &ErrorDetail{
		Code:    code,
		Message: msg,
		Details: details,
	})
}

// FormatErrorText formats an error for text output
func FormatErrorText(msg string, suggestion string) string {
	output := fmt.Sprintf("Error: %s", msg)
	if suggestion != "" {
		output += fmt.Sprintf("\nSuggestion: %s", suggestion)
	}
	return output
}

func formatJSON(data interface{}, success bool, message string, errDetail *ErrorDetail) string {
	resp := Response{
		Success: success,
		Data:    data,
		Message: message,
		Error:   errDetail,
		Metadata: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Remove empty fields
	if message == "" {
		resp.Message = ""
	}
	if data == nil && success {
		resp.Data = nil
	}

	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"success": false, "error": {"code": "MARSHAL_ERROR", "message": %q}}`, err.Error())
	}
	return string(b)
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
		if len(keys) >= 6 { // Limit columns
			break
		}
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
	maxRows := 50
	for i, item := range data {
		if i >= maxRows {
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
			if len(str) > 30 {
				str = str[:27] + "..."
			}
			values = append(values, str)
		}
		lines = append(lines, strings.Join(values, " | "))
	}

	if len(data) > maxRows {
		lines = append(lines, fmt.Sprintf("... and %d more items", len(data)-maxRows))
	}

	return strings.Join(lines, "\n")
}

func formatDict(data map[string]interface{}) string {
	var lines []string

	for key, value := range data {
		keyLabel := formatKey(key)

		switch v := value.(type) {
		case map[string]interface{}:
			lines = append(lines, fmt.Sprintf("%s:", keyLabel))
			for k, val := range v {
				lines = append(lines, fmt.Sprintf("  %s: %s", k, formatValue(val)))
			}
		case []interface{}:
			lines = append(lines, fmt.Sprintf("%s:", keyLabel))
			maxItems := 10
			for i, item := range v {
				if i >= maxItems {
					lines = append(lines, fmt.Sprintf("  ... and %d more", len(v)-maxItems))
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

// PrintOutput prints formatted output to stdout
func PrintOutput(data interface{}, textMode bool, message string) {
	output := FormatOutput(data, textMode, message)
	fmt.Println(output)
}

// PrintSuccess prints a successful response
func PrintSuccess(data interface{}, textMode bool, message string) {
	if textMode {
		fmt.Println(formatText(data, message))
	} else {
		fmt.Println(FormatSuccess(data, message))
	}
}

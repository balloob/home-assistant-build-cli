// Package input parses user-supplied data from files, inline flags, or stdin
// in JSON or YAML format.
package input

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ghodss/yaml"
)

// ParseInput reads and parses input from various sources
func ParseInput(data, file, format string) (map[string]interface{}, error) {
	var inputData []byte
	var err error

	if file != "" {
		// Read from file
		inputData, err = os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		// Auto-detect format from extension
		if format == "" {
			if strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
				format = "yaml"
			} else if strings.HasSuffix(file, ".json") {
				format = "json"
			}
		}
	} else if data != "" {
		// Use provided data string
		inputData = []byte(data)
	} else {
		// Read from stdin
		inputData, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
	}

	if len(inputData) == 0 {
		return nil, fmt.Errorf("no input data provided")
	}

	// Auto-detect format if not specified
	if format == "" {
		trimmed := strings.TrimSpace(string(inputData))
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			format = "json"
		} else {
			format = "yaml"
		}
	}

	var result map[string]interface{}

	switch format {
	case "json":
		if err := json.Unmarshal(inputData, &result); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	case "yaml":
		// Convert YAML to JSON first, then to map
		jsonData, err := yaml.YAMLToJSON(inputData)
		if err != nil {
			return nil, fmt.Errorf("invalid YAML: %w", err)
		}
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return result, nil
}

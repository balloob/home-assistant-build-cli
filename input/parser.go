// Package input parses user-supplied data from files, inline flags, or stdin
// in JSON or YAML format.
package input

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ghodss/yaml"
)

// ParseInput reads and parses input from various sources
func ParseInput(data, file, format string) (map[string]interface{}, error) {
	return parseInput(data, file, format, false)
}

// ParseInputExact retains JSON number precision for read-modify-write workflows.
// The legacy ParseInput keeps its float64 behavior for existing consumers.
func ParseInputExact(data, file, format string) (map[string]interface{}, error) {
	return parseInput(data, file, format, true)
}

func parseInput(data, file, format string, exactNumbers bool) (map[string]interface{}, error) {
	var inputData []byte
	var err error

	if data != "" && file != "" {
		return nil, fmt.Errorf("conflicting input sources: use either --data or --file")
	}

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
		format = "yaml" // default
		for _, b := range inputData {
			switch b {
			case ' ', '\t', '\n', '\r':
				continue
			case '{', '[':
				format = "json"
			}
			break
		}
	}

	var result map[string]interface{}
	if exactNumbers {
		if format == "yaml" {
			inputData, err = yaml.YAMLToJSON(inputData)
			if err != nil {
				return nil, fmt.Errorf("invalid YAML: %w", err)
			}
		} else if format != "json" {
			return nil, fmt.Errorf("unsupported format: %s", format)
		}
		if err := DecodeJSONExact(inputData, &result); err != nil {
			return nil, err
		}
		return result, nil
	}

	switch format {
	case "json":
		if err := json.Unmarshal(inputData, &result); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(inputData, &result); err != nil {
			return nil, fmt.Errorf("invalid YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return result, nil
}

// DecodeJSONExact rejects trailing input and preserves number tokens.
func DecodeJSONExact(data []byte, result any) error {
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(result)
}

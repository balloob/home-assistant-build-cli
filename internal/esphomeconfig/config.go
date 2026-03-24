package esphomeconfig

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/input"
	"gopkg.in/yaml.v3"
)

type patchAssignment struct {
	Path  string
	Value any
}

type yamlPathElement struct {
	Key   string
	Index int
	IsMap bool
}

type taggedYAMLNode struct {
	Path  []yamlPathElement
	Kind  yaml.Kind
	Tag   string
	Value string
}

// ParseConfig parses YAML content into a generic object map.
func ParseConfig(content string) (map[string]any, error) {
	if strings.TrimSpace(content) == "" {
		return map[string]any{}, nil
	}

	value, err := input.ParseValue(content)
	if err != nil {
		return nil, err
	}

	config, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("ESPHome config must be a YAML object")
	}

	return config, nil
}

// RenderConfig renders a generic object map into YAML.
func RenderConfig(config map[string]any) (string, error) {
	return input.MarshalYAML(config)
}

// ApplyPatch overlays object data and set assignments on top of YAML content.
func ApplyPatch(content string, overlay map[string]any, setValues []string) (string, error) {
	taggedNodes, err := collectTaggedYAMLNodes(content)
	if err != nil {
		return "", err
	}

	config, err := ParseConfig(content)
	if err != nil {
		return "", err
	}

	if overlay != nil {
		config = input.MergeMaps(config, overlay)
	}

	assignments, err := parsePatchAssignments(setValues)
	if err != nil {
		return "", err
	}

	for _, assignment := range assignments {
		if err := input.SetPathValue(config, assignment.Path, assignment.Value); err != nil {
			return "", fmt.Errorf("apply patch %q: %w", assignment.Path, err)
		}
	}

	rendered, err := RenderConfig(config)
	if err != nil {
		return "", err
	}

	if len(taggedNodes) == 0 {
		return rendered, nil
	}

	return restoreTaggedYAMLNodes(rendered, taggedNodes)
}

// ExtractCreateDetails returns useful values from generated YAML.
func ExtractCreateDetails(content string) map[string]any {
	config, err := ParseConfig(content)
	if err != nil {
		return map[string]any{"content": content}
	}

	result := map[string]any{"content": content}
	if api, ok := config["api"].(map[string]any); ok {
		if encryption, ok := api["encryption"].(map[string]any); ok {
			if key, ok := encryption["key"].(string); ok && key != "" {
				result["api_encryption_key"] = key
			}
		}
	}
	if otaPassword := extractOTAPassword(config["ota"]); otaPassword != "" {
		result["ota_password"] = otaPassword
	}

	return result
}

// NormalizeWizardPlatform maps hab platform values to ESPHome wizard platform values.
func NormalizeWizardPlatform(platform string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "esp32":
		return "ESP32", nil
	case "esp8266":
		return "ESP8266", nil
	case "rp2040":
		return "RP2040", nil
	case "bk72xx":
		return "BK72XX", nil
	case "ln882x":
		return "LN882X", nil
	case "rtl87xx":
		return "RTL87XX", nil
	default:
		return "", fmt.Errorf("unsupported ESPHome platform %q", platform)
	}
}

func parsePatchAssignments(values []string) ([]patchAssignment, error) {
	assignments := make([]patchAssignment, 0, len(values))
	for _, raw := range values {
		path, valueText, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --set value %q (expected path=value)", raw)
		}
		value, err := input.ParseValue(valueText)
		if err != nil {
			return nil, fmt.Errorf("parse --set value for %q: %w", path, err)
		}
		assignments = append(assignments, patchAssignment{
			Path:  strings.TrimSpace(path),
			Value: value,
		})
	}

	return assignments, nil
}

func extractOTAPassword(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		password, _ := typed["password"].(string)
		return password
	case []any:
		for _, item := range typed {
			password := extractOTAPassword(item)
			if password != "" {
				return password
			}
		}
	}

	return ""
}

func collectTaggedYAMLNodes(content string) ([]taggedYAMLNode, error) {
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}

	var document yaml.Node
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		return nil, fmt.Errorf("parse YAML for tag preservation: %w", err)
	}

	if len(document.Content) == 0 {
		return nil, nil
	}

	result := make([]taggedYAMLNode, 0)
	collectTaggedYAMLNodesRecursive(document.Content[0], nil, &result)
	return result, nil
}

func collectTaggedYAMLNodesRecursive(node *yaml.Node, path []yamlPathElement, result *[]taggedYAMLNode) {
	if node == nil {
		return
	}

	if shouldPreserveTag(node.Tag) {
		elementPath := append([]yamlPathElement(nil), path...)
		*result = append(*result, taggedYAMLNode{
			Path:  elementPath,
			Kind:  node.Kind,
			Tag:   node.Tag,
			Value: node.Value,
		})
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			collectTaggedYAMLNodesRecursive(value, append(path, yamlPathElement{Key: key.Value, IsMap: true}), result)
		}
	case yaml.SequenceNode:
		for i, item := range node.Content {
			collectTaggedYAMLNodesRecursive(item, append(path, yamlPathElement{Index: i}), result)
		}
	}
}

func shouldPreserveTag(tag string) bool {
	return strings.HasPrefix(tag, "!") && !strings.HasPrefix(tag, "!!")
}

func restoreTaggedYAMLNodes(content string, taggedNodes []taggedYAMLNode) (string, error) {
	if len(taggedNodes) == 0 {
		return content, nil
	}

	var document yaml.Node
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		return "", fmt.Errorf("parse patched YAML for tag restoration: %w", err)
	}

	if len(document.Content) == 0 {
		return content, nil
	}

	for _, tagged := range taggedNodes {
		target := findYAMLNodeByPath(document.Content[0], tagged.Path)
		if target == nil || target.Kind != tagged.Kind {
			continue
		}
		if target.Kind == yaml.ScalarNode && target.Value != tagged.Value {
			continue
		}
		target.Tag = tagged.Tag
	}

	restored, err := yaml.Marshal(&document)
	if err != nil {
		return "", fmt.Errorf("marshal YAML with restored tags: %w", err)
	}

	return string(restored), nil
}

func findYAMLNodeByPath(root *yaml.Node, path []yamlPathElement) *yaml.Node {
	current := root
	for _, element := range path {
		if element.IsMap {
			if current.Kind != yaml.MappingNode {
				return nil
			}

			var next *yaml.Node
			for i := 0; i+1 < len(current.Content); i += 2 {
				if current.Content[i].Value == element.Key {
					next = current.Content[i+1]
					break
				}
			}
			if next == nil {
				return nil
			}
			current = next
			continue
		}

		if current.Kind != yaml.SequenceNode {
			return nil
		}
		if element.Index < 0 || element.Index >= len(current.Content) {
			return nil
		}
		current = current.Content[element.Index]
	}

	return current
}

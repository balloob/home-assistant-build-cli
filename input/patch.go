package input

import (
	"fmt"
	"strconv"
	"strings"

	ghodssyaml "github.com/ghodss/yaml"
	"gopkg.in/yaml.v3"
)

type patchPathToken struct {
	Key   string
	Index *int
}

// ParseValue parses a YAML or JSON scalar, list, or object value.
func ParseValue(raw string) (any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	var value any
	if err := yaml.Unmarshal([]byte(raw), &value); err != nil {
		return raw, nil
	}
	return normalizeValue(value), nil
}

// MergeMaps recursively overlays src onto dst.
func MergeMaps(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = map[string]any{}
	}
	for key, value := range src {
		srcMap, srcIsMap := value.(map[string]any)
		dstMap, dstIsMap := dst[key].(map[string]any)
		if srcIsMap && dstIsMap {
			dst[key] = MergeMaps(dstMap, srcMap)
			continue
		}
		dst[key] = value
	}
	return dst
}

// SetPathValue sets a value using dotted keys and optional list indexes.
func SetPathValue(root map[string]any, path string, value any) error {
	tokens, err := parsePatchPath(path)
	if err != nil {
		return err
	}
	if tokens[0].Index != nil {
		return fmt.Errorf("path %q cannot start with a list index", path)
	}

	updated, err := setPathValue(root, tokens, value)
	if err != nil {
		return err
	}
	if _, ok := updated.(map[string]any); !ok {
		return fmt.Errorf("path %q did not resolve to an object", path)
	}
	return nil
}

// MarshalYAML renders a generic structure as YAML.
func MarshalYAML(value any) (string, error) {
	data, err := ghodssyaml.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parsePatchPath(path string) ([]patchPathToken, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return nil, fmt.Errorf("path is required")
	}

	tokens := make([]patchPathToken, 0)
	var current strings.Builder
	flushKey := func() {
		if current.Len() == 0 {
			return
		}
		text := current.String()
		current.Reset()
		if idx, err := strconv.Atoi(text); err == nil {
			tokens = append(tokens, patchPathToken{Index: &idx})
			return
		}
		tokens = append(tokens, patchPathToken{Key: text})
	}

	for i := 0; i < len(trimmed); i++ {
		switch trimmed[i] {
		case '.':
			flushKey()
		case '[':
			flushKey()
			end := strings.IndexByte(trimmed[i:], ']')
			if end < 0 {
				return nil, fmt.Errorf("path %q has an unterminated list index", path)
			}
			indexText := trimmed[i+1 : i+end]
			index, err := strconv.Atoi(indexText)
			if err != nil {
				return nil, fmt.Errorf("path %q has an invalid list index %q", path, indexText)
			}
			tokens = append(tokens, patchPathToken{Index: &index})
			i += end
		default:
			current.WriteByte(trimmed[i])
		}
	}
	flushKey()

	if len(tokens) == 0 {
		return nil, fmt.Errorf("path is required")
	}
	return tokens, nil
}

func setPathValue(container any, tokens []patchPathToken, value any) (any, error) {
	if len(tokens) == 0 {
		return value, nil
	}

	token := tokens[0]
	if token.Index != nil {
		items, err := ensureSlice(container)
		if err != nil {
			return nil, err
		}
		idx := *token.Index
		for len(items) <= idx {
			items = append(items, nil)
		}
		updated, err := setPathValue(items[idx], tokens[1:], value)
		if err != nil {
			return nil, err
		}
		items[idx] = updated
		return items, nil
	}

	objects, err := ensureMap(container)
	if err != nil {
		return nil, err
	}
	updated, err := setPathValue(objects[token.Key], tokens[1:], value)
	if err != nil {
		return nil, err
	}
	objects[token.Key] = updated
	return objects, nil
}

func ensureMap(value any) (map[string]any, error) {
	if value == nil {
		return map[string]any{}, nil
	}
	objects, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot descend into non-object value")
	}
	return objects, nil
}

func ensureSlice(value any) ([]any, error) {
	if value == nil {
		return []any{}, nil
	}
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("cannot descend into non-list value")
	}
	return items, nil
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		normalized := make(map[string]any, len(typed))
		for key, item := range typed {
			normalized[key] = normalizeValue(item)
		}
		return normalized
	case map[any]any:
		normalized := make(map[string]any, len(typed))
		for key, item := range typed {
			normalized[fmt.Sprint(key)] = normalizeValue(item)
		}
		return normalized
	case []any:
		for i, item := range typed {
			typed[i] = normalizeValue(item)
		}
		return typed
	default:
		return typed
	}
}

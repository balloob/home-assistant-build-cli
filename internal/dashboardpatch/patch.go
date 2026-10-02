// Package dashboardpatch implements optimistic, field-level dashboard changes.
// It has no CLI or transport dependencies.
package dashboardpatch

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/input"
)

type Edit struct {
	Path   string
	Value  any
	Remove bool
}

type Spec struct {
	Target string
	Merge  map[string]any
	Edits  []Edit
}

type Change struct {
	Path         string `json:"path"`
	Before       any    `json:"before"`
	After        any    `json:"after"`
	BeforeExists bool   `json:"before_exists"`
	AfterExists  bool   `json:"after_exists"`
}

// Revision hashes the entire JSON document, including fields outside the target.
func Revision(config map[string]any) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("encode dashboard: %w", err)
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data)), nil
}

func clone[T any](config T) (T, error) {
	var result T
	data, err := json.Marshal(config)
	if err != nil {
		return result, err
	}
	if err := input.DecodeJSONExact(data, &result); err != nil {
		return result, err
	}
	return result, nil
}

// Apply preserves the input, deep-merges objects, and replaces explicit array
// values atomically. JSON pointers never implicitly create intermediate paths.
func Apply(before map[string]any, spec Spec) (map[string]any, []Change, error) {
	normalizedBefore, err := clone(before)
	if err != nil {
		return nil, nil, err
	}
	after, err := clone(normalizedBefore)
	if err != nil {
		return nil, nil, err
	}
	if after == nil {
		return nil, nil, fmt.Errorf("dashboard config must be an object")
	}
	tokens, err := pointerTokens(spec.Target)
	if err != nil {
		return nil, nil, err
	}
	selected, err := resolve(after, tokens)
	if err != nil {
		return nil, nil, err
	}
	target, ok := selected.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("target must resolve to an object")
	}
	overlay, err := clone(spec.Merge)
	if err != nil {
		return nil, nil, err
	}
	input.MergeMaps(target, overlay)
	for _, edit := range spec.Edits {
		edit.Value, err = clone(edit.Value)
		if err != nil {
			return nil, nil, err
		}
		if err := applyEdit(target, edit); err != nil {
			return nil, nil, err
		}
	}
	changes := []Change{}
	diff("", normalizedBefore, after, true, true, &changes)
	return after, changes, nil
}

func pointerTokens(pointer string) ([]string, error) {
	if pointer == "" {
		return nil, nil
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("invalid JSON pointer: must start with /")
	}
	tokens := strings.Split(pointer[1:], "/")
	for i, token := range tokens {
		for j := 0; j < len(token); j++ {
			if token[j] == '~' {
				if j+1 == len(token) || (token[j+1] != '0' && token[j+1] != '1') {
					return nil, fmt.Errorf("invalid JSON pointer escape")
				}
				j++
			}
		}
		tokens[i] = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
	}
	return tokens, nil
}

func arrayIndex(key string, length int) (int, error) {
	i, err := strconv.Atoi(key)
	if err != nil || i < 0 || i >= length || strconv.Itoa(i) != key {
		return 0, fmt.Errorf("invalid or out-of-range array index")
	}
	return i, nil
}

func resolve(value any, tokens []string) (any, error) {
	for _, token := range tokens {
		switch parent := value.(type) {
		case map[string]any:
			var ok bool
			value, ok = parent[token]
			if !ok {
				return nil, fmt.Errorf("JSON pointer does not exist")
			}
		case []any:
			i, err := arrayIndex(token, len(parent))
			if err != nil {
				return nil, err
			}
			value = parent[i]
		default:
			return nil, fmt.Errorf("JSON pointer descends into a scalar")
		}
	}
	return value, nil
}

func applyEdit(target map[string]any, edit Edit) error {
	tokens, err := pointerTokens(edit.Path)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return fmt.Errorf("edit must name a field; root replacement is not supported")
	}
	parent, err := resolve(target, tokens[:len(tokens)-1])
	if err != nil {
		return err
	}
	key := tokens[len(tokens)-1]
	switch parent := parent.(type) {
	case map[string]any:
		if edit.Remove {
			delete(parent, key)
		} else {
			parent[key] = edit.Value
		}
	case []any:
		if edit.Remove {
			return fmt.Errorf("remove array elements by replacing the containing array explicitly")
		}
		i, err := arrayIndex(key, len(parent))
		if err != nil {
			return err
		}
		parent[i] = edit.Value
	default:
		return fmt.Errorf("edit parent must be an object or array")
	}
	return nil
}

func escape(token string) string {
	return strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1")
}

func diff(path string, before, after any, beforeExists, afterExists bool, changes *[]Change) {
	if beforeExists == afterExists && reflect.DeepEqual(before, after) {
		return
	}
	b, bm := before.(map[string]any)
	a, am := after.(map[string]any)
	if bm && am {
		keys := make([]string, 0, len(b)+len(a))
		for k := range b {
			keys = append(keys, k)
		}
		for k := range a {
			if _, ok := b[k]; !ok {
				keys = append(keys, k)
			}
		}
		slices.Sort(keys)
		for _, key := range keys {
			bv, be := b[key]
			av, ae := a[key]
			diff(path+"/"+escape(key), bv, av, be, ae, changes)
		}
		return
	}
	bs, ba := before.([]any)
	as, aa := after.([]any)
	if ba && aa && len(bs) == len(as) {
		for i := range bs {
			diff(path+"/"+strconv.Itoa(i), bs[i], as[i], true, true, changes)
		}
		return
	}
	*changes = append(*changes, Change{path, before, after, beforeExists, afterExists})
}

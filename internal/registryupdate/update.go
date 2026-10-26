// Package registryupdate previews and verifies registry metadata updates.
package registryupdate

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

// Store provides access to one kind of registry resource.
type Store interface {
	Get(string) (map[string]any, error)
	Update(string, map[string]any) (map[string]any, error)
}

type Change struct {
	Before any `json:"before"`
	After  any `json:"after"`
}

type Result struct {
	Mode        string            `json:"mode"`
	Changes     map[string]Change `json:"changes"`
	WouldChange bool              `json:"would_change"`
	Saved       *bool             `json:"saved"`
	Verified    bool              `json:"verified"`
	Entry       map[string]any    `json:"entry,omitempty"`
}

// Output preserves the resource fields returned by older update commands.
func (r Result) Output() map[string]any {
	out := maps.Clone(r.Entry)
	if out == nil {
		out = map[string]any{}
	}
	out["mode"], out["changes"], out["would_change"] = r.Mode, r.Changes, r.WouldChange
	out["saved"], out["verified"] = r.Saved, r.Verified
	return out
}

// Run reads current values before planning, and independently reads back writes.
// A failed write acknowledgement is never retried automatically.
func Run(store Store, id string, params map[string]any, plan bool) (Result, error) {
	saved := false
	result := Result{Mode: "apply", Changes: map[string]Change{}, Saved: &saved}
	if plan {
		result.Mode = "plan"
	}
	if len(params) == 0 {
		return result, fmt.Errorf("no update parameters provided")
	}
	entry, err := store.Get(id)
	if err != nil {
		return result, fmt.Errorf("read registry entry: %w", err)
	}
	result.Entry = entry
	for key, value := range params {
		before, exists := entry[key]
		if !exists {
			return result, fmt.Errorf("registry entry does not expose %s; check Home Assistant version support", key)
		}
		result.Changes[key] = Change{Before: before, After: value}
		result.WouldChange = result.WouldChange || !equal(key, before, value)
	}
	if plan {
		return result, nil
	}
	if !result.WouldChange {
		result.Verified = true
		return result, nil
	}
	if _, err := store.Update(id, params); err != nil {
		result.Saved = nil
		return result, fmt.Errorf("update was not acknowledged; inspect the registry before retrying: %w", err)
	}
	*result.Saved = true
	entry, err = store.Get(id)
	if err != nil {
		return result, fmt.Errorf("update was saved but read-back failed: %w", err)
	}
	result.Entry = entry
	for key, change := range result.Changes {
		value, exists := entry[key]
		if !exists || !equal(key, value, change.After) {
			return result, fmt.Errorf("update was saved but read-back did not confirm %s", key)
		}
	}
	result.Verified = true
	return result, nil
}

func equal(key string, a, b any) bool {
	if key == "aliases" || key == "labels" {
		return reflect.DeepEqual(stringSet(a), stringSet(b))
	}
	// Registry JSON decodes numbers as float64; CLI flags supply integers.
	left, err := json.Marshal(a)
	if err != nil {
		return false
	}
	right, err := json.Marshal(b)
	return err == nil && string(left) == string(right)
}

func stringSet(value any) []string {
	values := []string{}
	switch v := value.(type) {
	case []string:
		values = append(values, v...)
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				values = append(values, s)
			}
		}
	}
	slices.Sort(values)
	return slices.Compact(values)
}

// NormalizeSet matches HA's alias whitespace handling and set semantics.
func NormalizeSet(values []string, trim bool) []string {
	out := []string{}
	for _, value := range values {
		if trim {
			value = strings.TrimSpace(value)
		}
		if value != "" {
			out = append(out, value)
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// RequireIDs rejects unknown references, including labels HA silently drops.
func RequireIDs(list func() ([]any, error), key string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	entries, err := list()
	if err != nil {
		return fmt.Errorf("validate %s: %w", key, err)
	}
	for _, id := range ids {
		found := slices.ContainsFunc(entries, func(item any) bool {
			entry, ok := item.(map[string]any)
			return ok && entry[key] == id
		})
		if !found {
			return fmt.Errorf("unknown %s %q", key, id)
		}
	}
	return nil
}

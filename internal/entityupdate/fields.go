package entityupdate

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// stringList matches HA's unordered label sets and ordered, trimmed aliases.
func stringList(values []string, aliases bool) []any {
	if !aliases {
		values = slices.Clone(values)
		slices.Sort(values)
		values = slices.Compact(values)
	}
	result := make([]any, 0, len(values))
	for _, value := range values {
		if aliases {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
		}
		result = append(result, value)
	}
	return result
}

func normalizeLabels(value any) any {
	items, ok := value.([]any)
	if !ok {
		return value
	}
	labels := make([]string, 0, len(items))
	for _, item := range items {
		label, ok := item.(string)
		if !ok {
			return value
		}
		labels = append(labels, label)
	}
	return stringList(labels, false)
}

func plannedChanges(entry, params map[string]any) (map[string]Change, error) {
	changes := map[string]Change{}
	for key, value := range params {
		if key == "new_entity_id" {
			key = "entity_id"
		}
		if key == "categories" {
			current, ok := entry[key].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("source registry entry is missing categories")
			}
			merged := maps.Clone(current)
			for scope, id := range value.(map[string]any) {
				if id == nil {
					delete(merged, scope)
				} else {
					merged[scope] = id
				}
			}
			value = merged
		}
		changes[key] = Change{Before: entry[key], After: value}
	}
	return changes, nil
}

// checkReferences rejects unknown IDs before HA can silently drop labels or save
// other requested fields alongside an invalid assignment.
func checkReferences(store Store, opts Options) error {
	if opts.AreaID != nil && *opts.AreaID != "" {
		entries, err := store.AreaRegistryList()
		if err != nil {
			return fmt.Errorf("list areas: %w", err)
		}
		if err := requireRegistryIDs(entries, "area_id", []string{*opts.AreaID}); err != nil {
			return err
		}
	}
	if opts.Labels != nil && len(*opts.Labels) > 0 {
		entries, err := store.LabelRegistryList()
		if err != nil {
			return fmt.Errorf("list labels: %w", err)
		}
		if err := requireRegistryIDs(entries, "label_id", *opts.Labels); err != nil {
			return err
		}
	}
	for scope, id := range opts.Categories {
		if id == "" {
			continue
		}
		entries, err := store.CategoryRegistryList(scope)
		if err != nil {
			return fmt.Errorf("list categories for %s: %w", scope, err)
		}
		if err := requireRegistryIDs(entries, "category_id", []string{id}); err != nil {
			return fmt.Errorf("scope %s: %w", scope, err)
		}
	}
	return nil
}

func requireRegistryIDs(entries []any, key string, ids []string) error {
	known := map[string]bool{}
	for _, item := range entries {
		if entry, ok := item.(map[string]any); ok {
			if id, ok := entry[key].(string); ok {
				known[id] = true
			}
		}
	}
	for _, id := range ids {
		if !known[id] {
			return fmt.Errorf("unknown %s %q; no changes were made", key, id)
		}
	}
	return nil
}

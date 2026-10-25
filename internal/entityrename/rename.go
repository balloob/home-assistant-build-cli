// Package entityrename plans and verifies entity registry ID changes.
package entityrename

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Store is the registry, state, and relationship API needed for a rename.
type Store interface {
	EntityRegistryGet(string) (map[string]any, error)
	EntityRegistryList() ([]any, error)
	EntityRegistryUpdate(string, map[string]any) (map[string]any, error)
	GetStates() ([]any, error)
	SearchRelated(string, string) (map[string][]string, error)
}

type Result struct {
	Mode        string              `json:"mode"`
	OldEntityID string              `json:"old_entity_id"`
	NewEntityID string              `json:"new_entity_id"`
	UniqueID    string              `json:"unique_id"`
	Related     map[string][]string `json:"related"`
	Warnings    []string            `json:"warnings"`
	// Saved is nil when the write acknowledgement is uncertain.
	Saved    *bool `json:"saved"`
	Verified bool  `json:"verified"`
}

var entityIDPattern = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*\.[a-z0-9]+(_[a-z0-9]+)*$`)

// Validate rejects invalid IDs and domain changes before opening a connection.
func Validate(oldID, newID string) error {
	if !entityIDPattern.MatchString(oldID) || !entityIDPattern.MatchString(newID) {
		return fmt.Errorf("entity IDs must use domain.object_id with lowercase letters, numbers, and single underscores between words")
	}
	oldDomain, _, _ := strings.Cut(oldID, ".")
	newDomain, _, _ := strings.Cut(newID, ".")
	if oldDomain != newDomain {
		return fmt.Errorf("the new entity ID must keep the %s domain", oldDomain)
	}
	if oldID == newID {
		return fmt.Errorf("the new entity ID must differ from the current ID")
	}
	return nil
}

// Run always checks collisions and references before attempting a registry write.
// It never updates the integration's unique_id or rewrites references.
func Run(store Store, oldID, newID string, plan bool) (Result, error) {
	saved := false
	result := Result{Mode: "plan", OldEntityID: oldID, NewEntityID: newID, Saved: &saved}
	if !plan {
		result.Mode = "apply"
	}
	if err := Validate(oldID, newID); err != nil {
		return result, err
	}
	entry, err := store.EntityRegistryGet(oldID)
	if err != nil {
		return result, fmt.Errorf("read source registry entry: %w", err)
	}
	result.UniqueID, _ = entry["unique_id"].(string)
	if entry["entity_id"] != oldID || result.UniqueID == "" {
		return result, fmt.Errorf("source registry entry is missing its entity_id or unique_id")
	}
	if err := checkCollision(store, newID); err != nil {
		return result, err
	}
	result.Related, err = store.SearchRelated("entity", oldID)
	if err != nil {
		return result, fmt.Errorf("check related items: %w", err)
	}
	if result.Related == nil {
		result.Related = map[string][]string{}
	}
	for _, ids := range result.Related {
		slices.Sort(ids)
	}
	result.Warnings = []string{
		"References using the old entity ID may break. Related items come from search/related; hab does not rewrite them.",
		"Relationship search is not exhaustive: dashboards, templates, and dynamically constructed references are not covered. Inspect these manually even when related is empty.",
		"Long-term statistics use statistic_id. Home Assistant's recorder normally migrates history and statistics on registry renames, but existing target history/statistics can prevent migration. hab does not verify or repair that migration; check power/energy history afterwards.",
		"Only entity_id is changed; unique_id and the integration link are preserved.",
	}
	if plan {
		return result, nil
	}
	return apply(store, result)
}

func checkCollision(store Store, newID string) error {
	// Include disabled registry entries and live entities without registry entries.
	for _, read := range []func() ([]any, error){store.EntityRegistryList, store.GetStates} {
		entries, err := read()
		if err != nil {
			return fmt.Errorf("check entity ID collision: %w", err)
		}
		for _, item := range entries {
			if entry, ok := item.(map[string]any); ok && entry["entity_id"] == newID {
				return fmt.Errorf("target entity ID %s already exists; no changes were made", newID)
			}
		}
	}
	return nil
}

func apply(store Store, result Result) (Result, error) {
	result.Mode = "apply"
	_, err := store.EntityRegistryUpdate(result.OldEntityID, map[string]any{"new_entity_id": result.NewEntityID})
	if err != nil {
		// A transport error may occur after HA accepted the write. Do not retry.
		result.Saved = nil
		return result, fmt.Errorf("entity ID update was not acknowledged; inspect both IDs before retrying: %w", err)
	}
	*result.Saved = true
	entry, err := store.EntityRegistryGet(result.NewEntityID)
	if err != nil {
		return result, fmt.Errorf("entity ID was saved but read-back failed: %w", err)
	}
	if entry["entity_id"] != result.NewEntityID || entry["unique_id"] != result.UniqueID {
		return result, fmt.Errorf("entity ID was saved but read-back did not confirm the new ID and unchanged unique_id")
	}
	result.Verified = true
	return result, nil
}

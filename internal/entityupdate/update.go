// Package entityupdate plans and verifies entity registry changes.
package entityupdate

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

// Store is the registry, state, and relationship API needed for an update.
type Store interface {
	EntityRegistryGet(string) (map[string]any, error)
	EntityRegistryList() ([]any, error)
	EntityRegistryUpdate(string, map[string]any) (map[string]any, error)
	GetStates() ([]any, error)
	SearchRelated(string, string) (map[string][]string, error)
	AreaRegistryList() ([]any, error)
	LabelRegistryList() ([]any, error)
	CategoryRegistryList(string) ([]any, error)
}

// Options distinguishes omitted fields from explicit empty values or false.
type Options struct {
	EntityID    *string
	Name        *string
	Icon        *string
	AreaID      *string
	DeviceClass *string
	Labels      *[]string
	Aliases     *[]string
	Categories  map[string]string // Only these scopes are changed; empty IDs remove assignments.
	Hidden      *bool
	Disabled    *bool
}

type Change struct {
	Before any `json:"before"`
	After  any `json:"after"`
}

type Result struct {
	Mode        string              `json:"mode"`
	OldEntityID string              `json:"old_entity_id"`
	NewEntityID string              `json:"new_entity_id"`
	UniqueID    string              `json:"unique_id"`
	Changes     map[string]Change   `json:"changes"`
	Related     map[string][]string `json:"related"`
	Warnings    []string            `json:"warnings"`
	// Saved is nil when the write acknowledgement is uncertain.
	Saved    *bool `json:"saved"`
	Verified bool  `json:"verified"`
}

var entityIDPattern = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*\.[a-z0-9]+(_[a-z0-9]+)*$`)

// Validate rejects empty updates, invalid IDs, and domain changes before I/O.
func Validate(oldID string, opts Options) error {
	if opts.EntityID == nil && len(updateParams(oldID, opts)) == 0 {
		return fmt.Errorf("provide at least one registry field to update")
	}
	newID := oldID
	if opts.EntityID != nil {
		newID = *opts.EntityID
	}
	if !entityIDPattern.MatchString(oldID) || !entityIDPattern.MatchString(newID) {
		return fmt.Errorf("entity IDs must use domain.object_id with lowercase letters, numbers, and single underscores between words")
	}
	oldDomain, _, _ := strings.Cut(oldID, ".")
	newDomain, _, _ := strings.Cut(newID, ".")
	if oldDomain != newDomain {
		return fmt.Errorf("the new entity ID must keep the %s domain", oldDomain)
	}
	if oldID == newID && len(updateParams(oldID, opts)) == 0 {
		return fmt.Errorf("the new entity ID must differ from the current ID")
	}
	return nil
}

// Run checks collisions and references when an ID change is requested.
// It never updates the integration's unique_id or rewrites references.
func Run(store Store, oldID string, opts Options, plan bool) (Result, error) {
	saved := false
	result := Result{
		Mode: "plan", OldEntityID: oldID, NewEntityID: oldID, Saved: &saved,
		Related: map[string][]string{}, Changes: map[string]Change{},
		Warnings: []string{"unique_id and the integration link are preserved."},
	}
	if opts.EntityID != nil {
		result.NewEntityID = *opts.EntityID
	}
	if !plan {
		result.Mode = "apply"
	}
	if err := Validate(oldID, opts); err != nil {
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
	params := updateParams(oldID, opts)
	result.Changes, err = plannedChanges(entry, params)
	if err != nil {
		return result, err
	}
	if result.NewEntityID != oldID {
		if err := planRename(store, &result); err != nil {
			return result, err
		}
	}
	if err := checkReferences(store, opts); err != nil {
		return result, err
	}
	if opts.Disabled != nil {
		result.Warnings = append(result.Warnings, "Disabling removes the entity from active use; enabling may require a reload or restart. Verification checks registry settings, not runtime availability.")
	}
	if plan {
		return result, nil
	}
	return apply(store, result, params)
}

func updateParams(oldID string, opts Options) map[string]any {
	params := map[string]any{}
	if opts.EntityID != nil && *opts.EntityID != oldID {
		params["new_entity_id"] = *opts.EntityID
	}
	for key, value := range map[string]*string{"name": opts.Name, "icon": opts.Icon, "area_id": opts.AreaID, "device_class": opts.DeviceClass} {
		if value != nil {
			params[key] = nil
			if *value != "" {
				params[key] = *value
			}
		}
	}
	for key, value := range map[string]*bool{"hidden_by": opts.Hidden, "disabled_by": opts.Disabled} {
		if value != nil {
			params[key] = nil
			if *value {
				params[key] = "user"
			}
		}
	}
	if opts.Labels != nil {
		params["labels"] = stringList(*opts.Labels, false)
	}
	if opts.Aliases != nil {
		params["aliases"] = stringList(*opts.Aliases, true)
	}
	if len(opts.Categories) > 0 {
		categories := map[string]any{}
		for scope, id := range opts.Categories {
			categories[scope] = nil
			if id != "" {
				categories[scope] = id
			}
		}
		params["categories"] = categories
	}
	return params
}

func planRename(store Store, result *Result) error {
	if err := checkCollision(store, result.NewEntityID); err != nil {
		return err
	}
	related, err := store.SearchRelated("entity", result.OldEntityID)
	if err != nil {
		return fmt.Errorf("check related items: %w", err)
	}
	if related != nil {
		result.Related = related
	}
	for _, ids := range result.Related {
		slices.Sort(ids)
	}
	result.Warnings = append(result.Warnings,
		"References using the old entity ID may break. Related items come from search/related; hab does not rewrite them.",
		"Relationship search is not exhaustive: dashboards, templates, and dynamically constructed references are not covered. Inspect these manually even when related is empty.",
		"Long-term statistics use statistic_id. Home Assistant's recorder normally migrates history and statistics on registry renames, but existing target history/statistics can prevent migration. hab does not verify or repair that migration; check power/energy history afterwards.",
	)
	return nil
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

func apply(store Store, result Result, params map[string]any) (Result, error) {
	result.Mode = "apply"
	ack, err := store.EntityRegistryUpdate(result.OldEntityID, params)
	if err != nil {
		// A transport error may occur after HA accepted the write. Do not retry.
		result.Saved = nil
		return result, fmt.Errorf("entity update was not acknowledged; inspect the registry before retrying: %w", err)
	}
	*result.Saved = true
	if ack["require_restart"] == true {
		result.Warnings = append(result.Warnings, "Home Assistant reports that a restart is required to activate the entity.")
	} else if delay, ok := ack["reload_delay"].(float64); ok {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Home Assistant reports a configuration entry reload delay of %g seconds; runtime availability is not verified.", delay))
	}
	entry, err := store.EntityRegistryGet(result.NewEntityID)
	if err != nil {
		return result, fmt.Errorf("entity update was saved but read-back failed: %w", err)
	}
	if entry["entity_id"] != result.NewEntityID || entry["unique_id"] != result.UniqueID {
		return result, fmt.Errorf("entity update was saved but read-back did not confirm the expected entity_id and unchanged unique_id")
	}
	for key, change := range result.Changes {
		value, exists := entry[key]
		if key == "labels" {
			value = normalizeLabels(value)
		}
		if !exists || !reflect.DeepEqual(value, change.After) {
			return result, fmt.Errorf("entity update was saved but read-back did not confirm %s", key)
		}
	}
	result.Verified = true
	return result, nil
}

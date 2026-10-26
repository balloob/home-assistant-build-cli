package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/shortid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type idSelector struct {
	Argument  int
	Flag      string
	Namespace idNamespace
}

// prepareShortIDs runs before command-specific plans and confirmation prompts.
// Only audited ID selectors are rewritten, never arbitrary names or text.
func prepareShortIDs(cmd *cobra.Command, args []string) error {
	catalog := &idCatalog{}
	defer catalog.close()
	if err := resolveShortIDSelectors(cmd, args, catalog.list); err != nil {
		return err
	}
	for _, selector := range shortIDSelectors(cmd, args) {
		if selector.Namespace.Kind != "automation" && selector.Namespace.Kind != "scene" {
			continue
		}
		if selector.Flag != "" {
			continue
		}
		if selector.Argument >= len(args) {
			continue
		}
		full, err := configSelectorEntity(cmd, args[selector.Argument], selector.Namespace, catalog)
		if err != nil {
			return err
		}
		args[selector.Argument] = full
	}
	return nil
}

func resolveShortIDSelectors(cmd *cobra.Command, args []string, list func(idNamespace) ([]shortid.Candidate, error)) error {
	for _, selector := range shortIDSelectors(cmd, args) {
		if selector.Namespace.Kind == "trace" {
			selector.Namespace = shortIDNamespace("trace", cmd, args)
		}
		if selector.Flag != "" {
			flag := cmd.Flags().Lookup(selector.Flag)
			if flag == nil || !flag.Changed {
				continue
			}
			if slice, ok := flag.Value.(pflag.SliceValue); ok {
				values := slice.GetSlice()
				resolved := make([]string, len(values))
				for i, value := range values {
					full, err := resolveShortSelector(selector.Namespace, value, list)
					if err != nil {
						return err
					}
					resolved[i] = full
				}
				if err := slice.Replace(resolved); err != nil {
					return err
				}
				continue
			}
			full, err := resolveShortSelector(selector.Namespace, flag.Value.String(), list)
			if err != nil {
				return err
			}
			if err := flag.Value.Set(full); err != nil {
				return err
			}
		} else if selector.Argument < len(args) {
			full, err := resolveShortSelector(selector.Namespace, args[selector.Argument], list)
			if err != nil {
				return err
			}
			args[selector.Argument] = full
		}
	}
	return nil
}

func resolveShortSelector(ns idNamespace, raw string, list func(idNamespace) ([]shortid.Candidate, error)) (string, error) {
	if !shortid.Possible(raw) {
		return raw, nil
	}
	// Without a discovery window, send calendar UIDs unchanged. The server
	// matches them exactly; this preserves arbitrary integration-specific UIDs.
	// Prefix resolution requires the same explicit window as discovery.
	if ns.Kind == "calendar_event" && strings.HasSuffix(ns.Scope, "\n\n") {
		return raw, nil
	}
	candidates, err := list(ns)
	if err != nil {
		return "", err
	}
	full, err := shortid.Resolve(raw, candidates)
	if err != nil {
		var matchErr *shortid.Error
		if errors.As(err, &matchErr) {
			return "", &client.APIError{Code: "ID_RESOLUTION_FAILED", Message: err.Error(), Details: map[string]any{"resource": ns.Kind, "scope": ns.Scope, "input": raw, "candidates": matchErr.Candidates, "retryable": false}}
		}
		return "", err
	}
	if full != raw {
		noteResolvedShortID(ns, raw, full)
	}
	return full, nil
}

func noteResolvedShortID(ns idNamespace, input, full string) {
	executionMetadataMu.Lock()
	defer executionMetadataMu.Unlock()
	items, _ := executionMetadata["resolved_ids"].([]map[string]any)
	executionMetadata["resolved_ids"] = append(items, map[string]any{"resource": ns.Kind, "scope": ns.Scope, "input": input, "id": full})
}

var idArgumentKinds = map[string]string{
	"device_id": "device", "area_id": "area", "floor_id": "floor", "label_id": "label",
	"zone_id": "zone", "person_id": "person", "entry_id": "config_entry", "category_id": "category",
	"backup_id": "backup", "dataset_id": "thread", "notification_id": "notification",
	"dashboard_id": "dashboard", "item_uid": "todo_item", "uid": "calendar_event", "issue_id": "repair",
	"automation_id": "automation", "script_id": "script", "scene_id": "scene",
}
var idFlagKinds = map[string]string{
	"device": "device", "device-id": "device", "area": "area", "area-id": "area",
	"floor": "floor", "floor-id": "floor", "label": "label", "label-id": "label",
	"zone": "zone", "zone-id": "zone", "labels": "label", "agent": "backup_agent",
	"person": "person", "person-id": "person", "user-id": "user", "dataset": "thread", "dataset-id": "thread",
	"category": "category", "category-id": "category", "run-id": "trace",
	"automation": "automation", "automation-id": "automation", "script": "script", "script-id": "script", "scene": "scene",
}

func shortIDSelectors(cmd *cobra.Command, args []string) []idSelector {
	path := strings.TrimPrefix(schemaPath(cmd), "hab ")
	if strings.HasPrefix(path, "esphome ") || strings.HasPrefix(path, "marketplace ") {
		return nil
	}
	result := []idSelector{}
	for i, arg := range parseUseArgs(cmd.Use) {
		kind := idArgumentKinds[arg.Name]
		if arg.Name == "id" && strings.HasPrefix(path, "helper ") {
			kind = cmd.Annotations["hab.short_id_kind"]
		}
		// New resource identifiers are definitions, not selectors. Subcomponent
		// create commands still select an existing automation/script parent.
		if cmd.Name() == "create" && cmd.Parent() != nil && cmd.Parent().Parent() != nil && cmd.Parent().Parent().Parent() == nil {
			continue
		}
		if kind != "" {
			result = append(result, idSelector{Argument: i, Namespace: shortIDNamespace(kind, cmd, args)})
		}
	}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		kind := idFlagKinds[flag.Name]
		if kind != "" {
			result = append(result, idSelector{Flag: flag.Name, Namespace: shortIDNamespace(kind, cmd, args)})
		}
	})
	if path == "search related" {
		kind := commandIDValue(cmd, args, 0, "type", "item-type")
		if kind == "integration" {
			kind = ""
		} // integration is a domain here, not a config entry
		if kind != "" && supportedIDKind(kind) {
			ns := shortIDNamespace(kind, cmd, args)
			result = append(result, idSelector{Argument: 1, Namespace: ns}, idSelector{Flag: "id", Namespace: ns}, idSelector{Flag: "item-id", Namespace: ns})
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Namespace.Kind != "trace" && result[j].Namespace.Kind == "trace" })
	return result
}

func commandIDValue(cmd *cobra.Command, args []string, index int, flags ...string) string {
	for _, name := range flags {
		if flag := cmd.Flags().Lookup(name); flag != nil && flag.Changed {
			return flag.Value.String()
		}
	}
	if index >= 0 && index < len(args) {
		return args[index]
	}
	return ""
}
func shortIDNamespace(kind string, cmd *cobra.Command, args []string) idNamespace {
	ns := idNamespace{Kind: kind}
	switch kind {
	case "category":
		ns.Scope = commandIDValue(cmd, nil, -1, "scope")
		if ns.Scope == "" && cmd.Name() == "assign" {
			ns.Scope = inferCategoryScope(commandIDValue(cmd, args, 1, "entity", "entity-id"))
		}
	case "todo_item":
		ns.Scope = ensureDomainPrefix(commandIDValue(cmd, args, 0), "todo")
	case "calendar_event":
		ns.Scope = ensureDomainPrefix(commandIDValue(cmd, args, 0), "calendar") + "\n" + commandIDValue(cmd, nil, -1, "start") + "\n" + commandIDValue(cmd, nil, -1, "end")
	case "repair":
		ns.Scope = commandIDValue(cmd, args, 0)
	case "trace":
		ns.Scope = strings.TrimPrefix(commandIDValue(cmd, args, 0, "automation", "automation-id"), "automation.")
	}
	return ns
}
func supportedIDKind(kind string) bool {
	switch kind {
	case "zone", "device", "area", "floor", "label", "person", "config_entry", "category", "backup", "backup_agent", "thread", "user", "dashboard", "todo_item", "calendar_event", "notification", "repair", "trace", "automation", "scene", "script":
		return true
	}
	return strings.HasPrefix(kind, "helper:")
}

// resolveIDReferences resolves typed HA references in configuration input.
// Caller-owned definitions (id, unique_id, uid) and free text are not selectors.
func resolveIDReferences(data map[string]any) error {
	catalog := &idCatalog{}
	defer catalog.close()
	return walkIDReferences(data, catalog.list)
}
func walkIDReferences(value any, list func(idNamespace) ([]shortid.Candidate, error)) error {
	switch data := value.(type) {
	case map[string]any:
		for key, value := range data {
			if key == "variables" || key == "event_data" || key == "attributes" {
				continue
			}
			if key == "categories" {
				if categories, ok := value.(map[string]any); ok {
					for scope, id := range categories {
						if !validCategoryScopes[scope] {
							continue
						}
						resolved, err := resolveIDReferenceValue(id, idNamespace{Kind: "category", Scope: scope}, list)
						if err != nil {
							return fmt.Errorf("categories.%s: %w", scope, err)
						}
						categories[scope] = resolved
					}
				}
				continue
			}
			kind := idReferenceKinds[key]
			if kind != "" {
				resolved, err := resolveIDReferenceValue(value, idNamespace{Kind: kind}, list)
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				data[key] = resolved
			} else if err := walkIDReferences(value, list); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range data {
			if err := walkIDReferences(item, list); err != nil {
				return err
			}
		}
	}
	return nil
}
func resolveIDReferenceValue(value any, ns idNamespace, list func(idNamespace) ([]shortid.Candidate, error)) (any, error) {
	switch v := value.(type) {
	case string:
		return resolveShortSelector(ns, v, list)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			r, err := resolveIDReferenceValue(item, ns, list)
			if err != nil {
				return nil, err
			}
			out[i] = r
		}
		return out, nil
	}
	return value, nil
}

var idReferenceKinds = map[string]string{
	"device_id": "device", "area_id": "area", "floor_id": "floor", "label_id": "label", "user_id": "user",
	"config_entry_id": "config_entry", "backup_id": "backup", "agent_id": "backup_agent", "agent_ids": "backup_agent",
}

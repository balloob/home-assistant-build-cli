package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/shortid"
)

type idNamespace struct{ Kind, Scope string }
type idCatalog struct {
	ws    client.WebSocketAPI
	rest  client.RestAPI
	cache map[idNamespace][]shortid.Candidate
}

func (c *idCatalog) close() {
	if c.ws != nil {
		if err := c.ws.Close(); err != nil {
			noteWarning("ID discovery connection close failed.")
		}
	}
}
func (c *idCatalog) socket() (client.WebSocketAPI, error) {
	if c.ws == nil {
		ws, err := getWSClient()
		if err != nil {
			return nil, err
		}
		c.ws = ws
	}
	return c.ws, nil
}
func (c *idCatalog) http() (client.RestAPI, error) {
	if c.rest == nil {
		rest, err := getRESTClient()
		if err != nil {
			return nil, err
		}
		c.rest = rest
	}
	return c.rest, nil
}
func (c *idCatalog) list(ns idNamespace) ([]shortid.Candidate, error) {
	if entries, ok := c.cache[ns]; ok {
		return entries, nil
	}
	entries, key, err := c.entries(ns)
	if err != nil {
		return nil, fmt.Errorf("discover %s IDs: %w", ns.Kind, err)
	}
	candidates := shortid.Candidates(entries, key)
	if c.cache == nil {
		c.cache = map[idNamespace][]shortid.Candidate{}
	}
	c.cache[ns] = candidates
	return candidates, nil
}
func (c *idCatalog) entries(ns idNamespace) ([]any, string, error) {
	if ns.Kind == "calendar_event" {
		return c.calendar(ns.Scope)
	}
	ws, err := c.socket()
	if err != nil {
		return nil, "", err
	}
	var items []any
	key := "id"
	switch ns.Kind {
	case "zone":
		items, err = ws.ZoneList()
	case "device":
		items, err = ws.DeviceRegistryList()
	case "area":
		key = "area_id"
		items, err = ws.AreaRegistryList()
	case "floor":
		key = "floor_id"
		items, err = ws.FloorRegistryList()
	case "label":
		key = "label_id"
		items, err = ws.LabelRegistryList()
	case "person":
		items, err = ws.PersonRegistryList()
	case "config_entry":
		key = "entry_id"
		items, err = ws.ConfigEntriesList("")
	case "category":
		if !validCategoryScopes[ns.Scope] {
			return nil, "", fmt.Errorf("a valid category scope is required")
		}
		key = "category_id"
		items, err = ws.CategoryRegistryList(ns.Scope)
	case "todo_item":
		key = "uid"
		items, err = ws.TodoItemList(ns.Scope)
	case "notification":
		key = "notification_id"
		items, err = ws.NotificationList()
	case "backup":
		key = "backup_id"
		var data map[string]any
		data, err = ws.BackupInfo()
		items = objectList(data, "backups")
	case "backup_agent":
		key = "agent_id"
		var data map[string]any
		data, err = ws.BackupAgentsInfo()
		items = objectList(data, "agents")
	case "thread":
		key = "dataset_id"
		var data any
		data, err = ws.SendCommand("thread/list_datasets", nil)
		items = objectList(data, "datasets")
	case "user":
		var data any
		data, err = ws.SendCommand("config/auth/list", nil)
		items = objectList(data, "")
	case "dashboard":
		var data any
		data, err = ws.SendCommand("lovelace/dashboards/list", nil)
		items = objectList(data, "")
	case "repair":
		key = "issue_id"
		items, err = ws.RepairListIssues()
		items = filterIDEntries(items, "domain", ns.Scope)
	case "trace":
		key = "run_id"
		var data any
		scope := ns.Scope
		if shortid.Possible(scope) {
			scope, err = resolveShortSelector(idNamespace{Kind: "automation"}, scope, c.list)
			if err != nil {
				return nil, "", err
			}
		} else {
			rest, restErr := c.http()
			if restErr != nil {
				return nil, "", restErr
			}
			scope, err = resolveAutomationConfigID(rest, scope)
			if err != nil {
				return nil, "", err
			}
		}
		data, err = ws.SendCommand("trace/list", map[string]any{"domain": "automation", "item_id": scope})
		items = objectList(data, "")
	case "automation", "scene", "script":
		items, err = ws.GetStates()
		if err == nil {
			items, err = allConfigIDEntries(ws, items, ns.Kind)
		}
	default:
		if strings.HasPrefix(ns.Kind, "helper:") {
			items, err = ws.HelperList(strings.TrimPrefix(ns.Kind, "helper:"))
		} else {
			return nil, "", fmt.Errorf("unsupported ID namespace %s", ns.Kind)
		}
	}
	return items, key, err
}
func objectList(data any, key string) []any {
	if m, ok := data.(map[string]any); ok {
		data = m[key]
	}
	items, _ := data.([]any)
	return items
}
func filterIDEntries(items []any, key, value string) []any {
	result := []any{}
	for _, item := range items {
		if m, ok := item.(map[string]any); ok && m[key] == value {
			result = append(result, item)
		}
	}
	return result
}
func configIDEntries(states []any, domain string) []any {
	result := []any{}
	for _, item := range states {
		state, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entity, _ := state["entity_id"].(string)
		if !strings.HasPrefix(entity, domain+".") {
			continue
		}
		attrs, _ := state["attributes"].(map[string]any)
		id, _ := attrs["id"].(string)
		if domain == "script" {
			id = strings.TrimPrefix(entity, "script.")
		}
		if id != "" {
			result = append(result, map[string]any{"id": id, "name": attrs["friendly_name"], "entity_id": entity})
		}
	}
	return result
}

// Include registry entries without a current state (for example disabled
// automations), so they cannot be overlooked when checking prefix uniqueness.
func allConfigIDEntries(ws client.WebSocketAPI, states []any, domain string) ([]any, error) {
	registry, err := ws.EntityRegistryList()
	if err != nil {
		return nil, err
	}
	result := []any{}
	seen := map[string]bool{}
	seenEntities := map[string]bool{}
	for _, raw := range registry {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		entity, _ := entry["entity_id"].(string)
		id, _ := entry["unique_id"].(string)
		platform, _ := entry["platform"].(string)
		if !strings.HasPrefix(entity, domain+".") || id == "" || seen[id] {
			continue
		}
		if platform != domain && !(domain == "scene" && platform == "homeassistant") {
			continue
		}
		name, _ := entry["name"].(string)
		if name == "" {
			name, _ = entry["original_name"].(string)
		}
		if name == "" {
			name = entity
		}
		result = append(result, map[string]any{"id": id, "name": name, "entity_id": entity})
		seen[id] = true
		seenEntities[entity] = true
	}
	for _, raw := range configIDEntries(states, domain) {
		entry := raw.(map[string]any)
		id := entry["id"].(string)
		if !seen[id] && !seenEntities[entry["entity_id"].(string)] {
			result = append(result, entry)
			seen[id] = true
		}
	}
	return result, nil
}

func (c *idCatalog) calendar(scope string) ([]any, string, error) {
	// Calendar APIs can enumerate events only within a time window. The selected
	// window is part of the namespace and must be supplied for prefix deletion.
	parts := strings.Split(scope, "\n")
	if len(parts) != 3 || parts[1] == "" || parts[2] == "" {
		return nil, "", fmt.Errorf("calendar UID prefixes require --start and --end matching the discovery window")
	}
	rest, err := c.http()
	if err != nil {
		return nil, "", err
	}
	query := url.Values{"start": {parts[1]}, "end": {parts[2]}}
	data, err := rest.Get("calendars/" + url.PathEscape(parts[0]) + "?" + query.Encode())
	return objectList(data, ""), "uid", err
}

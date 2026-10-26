package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/internal/shortid"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

func init() {
	parent := &cobra.Command{Use: "id", Short: "Discover full and shortened resource IDs", GroupID: "registry"}
	var scope, start, end string
	list := &cobra.Command{
		Use: "list <resource>", Short: "List full IDs and their unambiguous abbreviations",
		Long: "List device, config_entry, category, user, thread, person, zone, area, floor, label,\nbackup, backup_agent, dashboard, notification, repair, trace, automation, script,\nscene, todo_item, or calendar_event IDs. --scope selects a category scope,\nentity ID for items/events, repair domain, or automation config ID for traces.\nCalendar events additionally require --start and --end.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind := args[0]
			if !supportedIDKind(kind) {
				return fmt.Errorf("unknown ID resource %q", kind)
			}
			ns := idNamespace{Kind: kind, Scope: scope}
			if kind == "calendar_event" {
				ns.Scope = scope + "\n" + start + "\n" + end
			}
			if (kind == "todo_item" || kind == "trace" || kind == "repair") && scope == "" {
				return fmt.Errorf("--scope is required for %s IDs", kind)
			}
			catalog := &idCatalog{}
			defer catalog.close()
			candidates, err := catalog.list(ns)
			if err != nil {
				return err
			}
			abbreviations := shortid.Abbreviations(candidates)
			items := make([]map[string]any, 0, len(candidates))
			for _, candidate := range candidates {
				item := map[string]any{"id": candidate.ID, "name": candidate.Name}
				if short, ok := abbreviations[candidate.ID]; ok {
					item["short_id"] = short
				}
				items = append(items, item)
			}
			output.PrintOutputWithContext(items, getTextMode(), "", output.EnvelopeContext{Operation: "list", ResourceType: kind})
			return nil
		},
	}
	list.Flags().StringVar(&scope, "scope", "", "Resource scope (category scope, entity ID, domain, or automation config ID)")
	list.Flags().StringVar(&start, "start", "", "Calendar discovery window start (ISO format)")
	list.Flags().StringVar(&end, "end", "", "Calendar discovery window end (ISO format)")
	mergeSchemaAnnotation(list, SchemaAnnotation{SideEffect: "read", Capabilities: []string{"auth", "ws", "rest"}, ResourceType: "resource_id", InputSources: []string{"args", "flags"}, GuideTopic: "discovery"})
	parent.AddCommand(list)
	rootCmd.AddCommand(parent)
}

func annotateIDs(items []any, key string) []any { return shortid.Annotate(items, key) }
func annotateIDResult(result any, container, key string) any {
	if items, ok := result.([]any); ok {
		return annotateIDs(items, key)
	}
	if m, ok := result.(map[string]any); ok {
		if items, ok := m[container].([]any); ok {
			m[container] = annotateIDs(items, key)
		}
	}
	return result
}
func displayID(item map[string]any, key string) string {
	if short, ok := item["short_id"].(string); ok && short != "" {
		return short
	}
	id, _ := item[key].(string)
	return id
}
func copyShortID(target, source map[string]any) map[string]any {
	if id, ok := source["config_id"].(string); ok {
		target["config_id"] = id
	}
	if short, ok := source["short_id"].(string); ok && short != "" {
		target["short_id"] = short
	}
	return target
}
func filterConfigEntries(items []any, domain string) []any {
	if domain == "" {
		return items
	}
	return filterIDEntries(items, "domain", domain)
}
func addConfigShortIDs(items []map[string]any, configs []any) {
	abbreviations := shortid.Abbreviations(shortid.Candidates(configs, "id"))
	byEntity := map[string]string{}
	for _, raw := range configs {
		m := raw.(map[string]any)
		byEntity[m["entity_id"].(string)] = m["id"].(string)
	}
	for _, item := range items {
		entity, _ := item["entity_id"].(string)
		id := byEntity[entity]
		if short, ok := abbreviations[id]; ok {
			item["config_id"] = id
			item["short_id"] = short
		}
	}
}

// Entity-facing operations still use entity IDs after a raw config ID has been
// selected. Configuration operations retain the full config ID.
func configSelectorEntity(cmd *cobra.Command, id string, ns idNamespace, catalog *idCatalog) (string, error) {
	path := strings.TrimPrefix(schemaPath(cmd), "hab ")
	if path != "automation run" && path != "scene activate" {
		return id, nil
	}
	if !shortid.Opaque(id) {
		return id, nil
	}
	entries, _, err := catalog.entries(ns)
	if err != nil {
		return "", err
	}
	for _, raw := range entries {
		m, ok := raw.(map[string]any)
		if ok && m["id"] == id {
			entity, _ := m["entity_id"].(string)
			if entity != "" {
				return entity, nil
			}
		}
	}
	return id, nil
}

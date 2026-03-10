package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
)

// stateListConfig holds the parameters for listDomainEntities.
type stateListConfig struct {
	// domain is the entity prefix to filter on (e.g. "automation", "script").
	domain string
	// listFlags is the ListFlags instance for --count/--brief/--limit handling.
	listFlags *ListFlags
	// textMode indicates whether text (vs JSON) output is requested.
	textMode bool
	// emptyMessage is printed in text mode when no items are found
	// (e.g. "No automations.").
	emptyMessage string
	// extraTextFields lists additional map keys to print as indented lines
	// in text mode (e.g. []string{"description", "blueprint"}).  Nil means
	// only the four base fields (alias, entity_id, state, last_triggered)
	// are printed.
	extraTextFields []string
	// enrichItems, if non-nil, is called after the initial state collection
	// to augment items with additional data (e.g. concurrent REST fetches).
	enrichItems func(items []map[string]interface{}) error
	// filterItem, if non-nil, is applied after enrichment.  Items for which
	// filterItem returns false are excluded from the result.
	filterItem func(item map[string]interface{}) bool
}

// listDomainEntities fetches all states via WebSocket, filters them by the
// configured domain prefix, and renders the result according to the ListFlags
// and text-mode settings.  Optional enrichment and filtering callbacks allow
// callers to add extended data or narrow down the result set.
func listDomainEntities(cfg stateListConfig) error {
	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	states, err := ws.GetStates()
	if err != nil {
		return err
	}

	prefix := cfg.domain + "."

	// Collect items matching the domain.
	items := make([]map[string]interface{}, 0, len(states)/4)
	for _, s := range states {
		state, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		entityID, _ := state["entity_id"].(string)
		if !strings.HasPrefix(entityID, prefix) {
			continue
		}
		attrs, _ := state["attributes"].(map[string]interface{})
		items = append(items, map[string]interface{}{
			"entity_id":      entityID,
			"alias":          attrs["friendly_name"],
			"state":          state["state"],
			"last_triggered": attrs["last_triggered"],
		})
	}

	// Optional enrichment (e.g. fetch descriptions/blueprints).
	if cfg.enrichItems != nil {
		if err := cfg.enrichItems(items); err != nil {
			return err
		}
	}

	// Optional filtering (e.g. blueprint filter).
	var result []map[string]interface{}
	if cfg.filterItem != nil {
		for _, item := range items {
			if cfg.filterItem(item) {
				result = append(result, item)
			}
		}
	} else {
		result = items
	}

	// ListFlags: --count, --limit, --brief.
	if cfg.listFlags.RenderCount(len(result), cfg.textMode) {
		return nil
	}
	result = cfg.listFlags.ApplyLimitMap(result)
	if cfg.listFlags.RenderBriefMap(result, cfg.textMode, "entity_id", "alias") {
		return nil
	}

	// Full output.
	if cfg.textMode {
		if len(result) == 0 {
			fmt.Println(cfg.emptyMessage)
			return nil
		}
		for _, item := range result {
			alias, _ := item["alias"].(string)
			entityID, _ := item["entity_id"].(string)
			state, _ := item["state"].(string)
			lastTriggered, _ := item["last_triggered"].(string)

			fmt.Printf("%s (%s): %s\n", alias, entityID, state)
			if lastTriggered != "" {
				fmt.Printf("  last_triggered: %s\n", lastTriggered)
			}
			for _, field := range cfg.extraTextFields {
				if val, _ := item[field].(string); val != "" {
					fmt.Printf("  %s: %s\n", field, val)
				}
			}
		}
	} else {
		output.PrintOutput(result, false, "")
	}
	return nil
}

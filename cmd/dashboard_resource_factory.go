package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ──────────────────────────────────────────────────────────
// Configuration types for the dashboard resource factory
// ──────────────────────────────────────────────────────────

// DashboardResourceFlag defines a custom flag to register on create/update commands.
type DashboardResourceFlag struct {
	Name      string // flag name (e.g. "title", "icon")
	Short     string // short flag (single char, or empty)
	Default   string // default value
	Usage     string // help text
	ConfigKey string // key to set in resource config (e.g. "title")
}

// DashboardResourceConfig describes a dashboard sub-resource (view, badge, section, card).
type DashboardResourceConfig struct {
	// ResourceName is the singular noun (e.g. "view", "badge", "section", "card").
	ResourceName string

	// ParentCmd is the dashboard command to attach this resource group to.
	ParentCmd *cobra.Command

	// GroupID is the cobra group ID for the parent command.
	GroupID string

	// ShortDesc / LongDesc for the parent group command.
	ShortDesc string
	LongDesc  string

	// ──── Nesting info ────

	// PathFromConfig describes how to navigate from the lovelace config root
	// to the array containing this resource. Each entry is an array key.
	// Examples:
	//   view:    ["views"]
	//   badge:   ["views", "badges"]
	//   section: ["views", "sections"]
	//   card:    ["views", "sections", "cards"]
	PathFromConfig []string

	// HasSectionFlag indicates that list/get/update/delete commands should accept
	// a --section flag to select which section (defaults to last).
	// Only applies to card.
	HasSectionFlag bool

	// ──── Get command style ────

	// GetUsesFlags indicates the get command uses --dashboard/--view/--index style
	// with positional fallback. If false, uses ExactArgs(N) positional only.
	GetUsesFlags bool

	// ──── Create overrides ────

	// CreateFlags are extra flags for the create command (beyond -d/--data, -f/--file, --format).
	CreateFlags []DashboardResourceFlag

	// CreateDefaults are key/value pairs to set on the resource config if not already present.
	// e.g. card: {"type": "tile"}, section: {"cards": []interface{}{}}
	CreateDefaults map[string]interface{}

	// CreateRequiredFields are keys that must be present after flag application.
	// e.g. view requires "title".
	CreateRequiredFields []string

	// CreateAutoScaffold enables auto-creation of parent views/sections if they don't exist.
	// Only used by card.
	CreateAutoScaffold bool

	// CreateViewArgOptional makes the view_index arg optional on create (card: RangeArgs(1,2)).
	CreateViewArgOptional bool

	// CreateLongDesc overrides the create command's Long description.
	CreateLongDesc string

	// CreateExample is the Cobra Example text for the create command.
	CreateExample string

	// ──── Update overrides ────

	// UpdateFlags are extra flags for the update command (beyond -d/--data, -f/--file, --format).
	UpdateFlags []DashboardResourceFlag

	// UpdateLongDesc overrides the update command's Long description.
	UpdateLongDesc string

	// ──── Badge-specific ────

	// ItemCanBeString indicates items in the resource array may be plain strings
	// (badges can be entity_id strings). Affects list/get/create/update.
	ItemCanBeString bool

	// ──── Delete overrides ────

	// DeleteLongDesc overrides the delete command's Long description.
	DeleteLongDesc string

	// ──── List overrides ────

	// ListLongDesc overrides the list command's Long description.
	ListLongDesc string
}

// ──────────────────────────────────────────────────────────
// Registration entry point
// ──────────────────────────────────────────────────────────

// RegisterDashboardResourceCRUD creates and registers a parent command with
// list, get, create, update, delete subcommands for a dashboard sub-resource.
func RegisterDashboardResourceCRUD(cfg DashboardResourceConfig) {
	name := cfg.ResourceName
	depth := len(cfg.PathFromConfig) // 1=view, 2=badge/section, 3=card

	// Parent group command
	parentCmd := &cobra.Command{
		Use:     name,
		Short:   cfg.ShortDesc,
		Long:    cfg.LongDesc,
		GroupID: cfg.GroupID,
	}
	cfg.ParentCmd.AddCommand(parentCmd)

	// Register subcommands
	registerDashResourceList(parentCmd, cfg, depth)
	registerDashResourceGet(parentCmd, cfg, depth)
	registerDashResourceCreate(parentCmd, cfg, depth)
	registerDashResourceUpdate(parentCmd, cfg, depth)
	registerDashResourceDelete(parentCmd, cfg, depth)
}

// ──────────────────────────────────────────────────────────
// Shared helpers
// ──────────────────────────────────────────────────────────

// fetchDashboardConfig fetches the lovelace config for the given dashboard URL path.
func fetchDashboardConfig(ws interface{ SendCommand(string, map[string]interface{}) (interface{}, error) }, urlPath string) (map[string]interface{}, error) {
	params := map[string]interface{}{}
	if urlPath != "lovelace" {
		params["url_path"] = urlPath
	}
	result, err := ws.SendCommand("lovelace/config", params)
	if err != nil {
		return nil, err
	}
	config, ok := result.(map[string]interface{})
	if !ok {
		return nil, nil // caller decides how to handle
	}
	return config, nil
}

// saveDashboardConfig saves the lovelace config for the given dashboard URL path.
func saveDashboardConfig(ws interface{ SendCommand(string, map[string]interface{}) (interface{}, error) }, urlPath string, config map[string]interface{}) error {
	saveParams := map[string]interface{}{
		"config": config,
	}
	if urlPath != "lovelace" {
		saveParams["url_path"] = urlPath
	}
	_, err := ws.SendCommand("lovelace/config/save", saveParams)
	return err
}

// confirmDelete prompts for deletion confirmation unless force or textMode is set.
func confirmDelete(force, textMode bool, description string) error {
	if !confirmAction(force, textMode, fmt.Sprintf("Are you sure you want to delete %s?", description)) {
		return fmt.Errorf("deletion cancelled")
	}
	return nil
}

// getViews extracts the views array from a dashboard config.
func getViews(config map[string]interface{}) ([]interface{}, error) {
	views, ok := config["views"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no views in dashboard")
	}
	return views, nil
}

// getViewAt returns the view map at the given index with bounds checking.
func getViewAt(views []interface{}, viewIndex int) (map[string]interface{}, error) {
	if viewIndex < 0 || viewIndex >= len(views) {
		return nil, fmt.Errorf("view index %d out of range (0-%d)", viewIndex, len(views)-1)
	}
	view, ok := views[viewIndex].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid view at index %d", viewIndex)
	}
	return view, nil
}

// getArrayFromMap extracts a []interface{} from a map by key.
func getArrayFromMap(m map[string]interface{}, key string) ([]interface{}, bool) {
	arr, ok := m[key].([]interface{})
	return arr, ok
}

// resolveSectionIndex resolves the section index, defaulting to the last section.
func resolveSectionIndex(sections []interface{}, sectionFlag int) (int, error) {
	if sections == nil || len(sections) == 0 {
		return -1, fmt.Errorf("no sections in view")
	}
	idx := sectionFlag
	if idx < 0 {
		idx = len(sections) - 1
	}
	if idx >= len(sections) {
		return -1, fmt.Errorf("section index %d out of range (0-%d)", idx, len(sections)-1)
	}
	return idx, nil
}

// getSectionAt returns the section map at the given index.
func getSectionAt(sections []interface{}, sectionIndex int) (map[string]interface{}, error) {
	section, ok := sections[sectionIndex].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid section at index %d", sectionIndex)
	}
	return section, nil
}

// badgeItemToMap normalizes a badge item (which can be a string or map) to a map.
func badgeItemToMap(item interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	switch v := item.(type) {
	case map[string]interface{}:
		for k, val := range v {
			data[k] = val
		}
	case string:
		data["entity"] = v
	}
	return data
}

// mapItemWithIndex copies a map item and adds an index field.
func mapItemWithIndex(item interface{}, index int, canBeString bool) map[string]interface{} {
	if canBeString {
		data := badgeItemToMap(item)
		data["index"] = index
		return data
	}
	data := make(map[string]interface{})
	if m, ok := item.(map[string]interface{}); ok {
		for k, val := range m {
			data[k] = val
		}
	}
	data["index"] = index
	return data
}

// ──────────────────────────────────────────────────────────
// Navigation: drill into the config tree to reach the resource array
// ──────────────────────────────────────────────────────────

// dashNavResult holds the navigation state after drilling into the config.
type dashNavResult struct {
	Config       map[string]interface{}   // root config
	Views        []interface{}            // config["views"]
	ViewIndex    int                      // resolved view index
	View         map[string]interface{}   // views[viewIndex]
	Sections     []interface{}            // view["sections"] (if depth >= 3)
	SectionIndex int                      // resolved section index (if depth >= 3)
	Section      map[string]interface{}   // sections[sectionIndex] (if depth >= 3)
	Items        []interface{}            // the target resource array
	ItemKey      string                   // the key in the parent map (e.g. "views", "badges", "sections", "cards")
	ParentMap    map[string]interface{}   // the map containing ItemKey
}

// navigateToResourceArray navigates from the root config to the resource array.
// args: [urlPath, viewIndex, sectionIndex?] as parsed integers where appropriate.
// For depth 1 (views): only config is needed.
// For depth 2 (badges/sections): needs viewIndex.
// For depth 3 (cards): needs viewIndex + sectionIndex.
func navigateToResourceArray(config map[string]interface{}, path []string, viewIndex, sectionIndex int) (*dashNavResult, error) {
	result := &dashNavResult{Config: config}
	result.ItemKey = path[len(path)-1]

	if len(path) >= 1 {
		views, err := getViews(config)
		if err != nil {
			return nil, err
		}
		result.Views = views

		if path[0] == "views" && len(path) == 1 {
			// Depth 1: views array
			result.Items = views
			result.ParentMap = config
			return result, nil
		}
	}

	// Depth >= 2: need viewIndex
	view, err := getViewAt(result.Views, viewIndex)
	if err != nil {
		return nil, err
	}
	result.ViewIndex = viewIndex
	result.View = view

	if len(path) == 2 {
		// Depth 2: badges or sections within a view
		items, ok := getArrayFromMap(view, path[1])
		if !ok {
			items = nil
		}
		result.Items = items
		result.ParentMap = view
		return result, nil
	}

	// Depth 3: cards within a section
	sections, ok := getArrayFromMap(view, path[1])
	if !ok {
		return nil, fmt.Errorf("no %s in view", path[1])
	}
	result.Sections = sections

	sIdx, err := resolveSectionIndex(sections, sectionIndex)
	if err != nil {
		return nil, err
	}
	result.SectionIndex = sIdx

	section, err := getSectionAt(sections, sIdx)
	if err != nil {
		return nil, err
	}
	result.Section = section

	items, _ := getArrayFromMap(section, path[2])
	result.Items = items
	result.ParentMap = section
	return result, nil
}

// saveBackToConfig writes the items array back through the navigation path and saves.
func saveBackToConfig(ws interface{ SendCommand(string, map[string]interface{}) (interface{}, error) }, urlPath string, nav *dashNavResult) error {
	nav.ParentMap[nav.ItemKey] = nav.Items

	// If depth 1 (views): Items IS the views array, so ParentMap[ItemKey]
	// already updated config["views"]. No further writes needed.
	if nav.ItemKey == "views" {
		return saveDashboardConfig(ws, urlPath, nav.Config)
	}

	// If depth >= 3, write section back
	if nav.Section != nil {
		nav.Sections[nav.SectionIndex] = nav.Section
		nav.View["sections"] = nav.Sections
	}
	// If depth >= 2, write view back
	if nav.View != nil {
		nav.Views[nav.ViewIndex] = nav.View
	}
	nav.Config["views"] = nav.Views

	return saveDashboardConfig(ws, urlPath, nav.Config)
}

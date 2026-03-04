package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/input"
	"github.com/home-assistant/hab/output"
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
	if !force && !textMode {
		fmt.Printf("Are you sure you want to delete %s? [y/N]: ", description)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
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

// ──────────────────────────────────────────────────────────
// LIST
// ──────────────────────────────────────────────────────────

func registerDashResourceList(parentCmd *cobra.Command, cfg DashboardResourceConfig, depth int) {
	name := cfg.ResourceName
	sectionFlag := -1

	// Determine arg count: depth 1 = 1 arg (urlPath), depth 2 = 2 args, depth 3 = 2 args (section via flag)
	argCount := depth
	if depth == 3 {
		argCount = 2 // urlPath + viewIndex; section is a flag
	}

	longDesc := fmt.Sprintf("List all %ss in a dashboard", name)
	if depth == 2 {
		longDesc = fmt.Sprintf("List all %ss in a dashboard view.", name)
	}
	if depth == 3 {
		longDesc = fmt.Sprintf("List all %ss in a dashboard section.\n\nIf section is not specified, uses the last section.", name)
	}
	if cfg.ListLongDesc != "" {
		longDesc = cfg.ListLongDesc
	}

	listUse := fmt.Sprintf("list <dashboard_url_path>")
	if depth >= 2 {
		listUse = "list <dashboard_url_path> <view_index>"
	}

	listCmd := &cobra.Command{
		Use:   listUse,
		Short: fmt.Sprintf("List %ss in a %s", name, listParentNoun(depth)),
		Long:  longDesc,
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			urlPath := args[0]
			viewIndex := -1
			if depth >= 2 {
				var err error
				viewIndex, err = strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid view index: %s", args[1])
				}
			}

			textMode := getTextMode()
			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			config, err := fetchDashboardConfig(ws, urlPath)
			if err != nil {
				return err
			}

			// For view list, handle nil config specially
			if depth == 1 && config == nil {
				output.PrintOutput([]interface{}{}, textMode, "")
				return nil
			}
			if config == nil {
				return fmt.Errorf("invalid dashboard config")
			}

			// For depth 1 (views), handle missing views gracefully
			if depth == 1 {
				views, ok := config["views"].([]interface{})
				if !ok {
					output.PrintOutput([]interface{}{}, textMode, "")
					return nil
				}
				itemList := make([]map[string]interface{}, len(views))
				for i, v := range views {
					itemList[i] = mapItemWithIndex(v, i, false)
				}
				output.PrintOutput(itemList, textMode, "")
				return nil
			}

			secIdx := sectionFlag
			if cfg.HasSectionFlag {
				secIdx, _ = cmd.Flags().GetInt("section")
			}
			nav, err := navigateToResourceArray(config, cfg.PathFromConfig, viewIndex, secIdx)
			if err != nil {
				return err
			}

			items := nav.Items
			if items == nil {
				output.PrintOutput([]interface{}{}, textMode, "")
				return nil
			}

			itemList := make([]map[string]interface{}, len(items))
			for i, item := range items {
				itemList[i] = mapItemWithIndex(item, i, cfg.ItemCanBeString)
			}

			output.PrintOutput(itemList, textMode, "")
			return nil
		},
	}

	if cfg.HasSectionFlag {
		listCmd.Flags().IntVarP(&sectionFlag, "section", "s", -1, "Section index (if "+name+"s are in a section)")
	}

	parentCmd.AddCommand(listCmd)
}

func listParentNoun(depth int) string {
	switch depth {
	case 1:
		return "dashboard"
	case 2:
		return "view"
	default:
		return "section"
	}
}

// ──────────────────────────────────────────────────────────
// GET
// ──────────────────────────────────────────────────────────

func registerDashResourceGet(parentCmd *cobra.Command, cfg DashboardResourceConfig, depth int) {
	name := cfg.ResourceName
	sectionFlag := -1

	if cfg.GetUsesFlags {
		registerDashResourceGetWithFlags(parentCmd, cfg, depth)
		return
	}

	// Positional-only get (badge style): ExactArgs(depth+1)
	argCount := depth + 1

	getUse := fmt.Sprintf("get <dashboard_url_path>")
	getUseParts := []string{"<dashboard_url_path>"}
	if depth >= 2 {
		getUseParts = append(getUseParts, "<view_index>")
	}
	if depth >= 3 {
		// For cards, the 3rd positional is card_index (section is a flag)
		getUseParts = append(getUseParts, fmt.Sprintf("<%s_index>", name))
		argCount = 3 // urlPath + viewIndex + cardIndex
	} else {
		getUseParts = append(getUseParts, fmt.Sprintf("<%s_index>", name))
	}
	getUse = "get " + strings.Join(getUseParts, " ")

	longDesc := fmt.Sprintf("Get a specific %s from a %s by index.", name, listParentNoun(depth))
	if depth == 3 {
		longDesc = fmt.Sprintf("Get a specific %s from a section by index.\n\nIf section is not specified, uses the last section.", name)
	}

	getCmd := &cobra.Command{
		Use:   getUse,
		Short: fmt.Sprintf("Get a specific %s", name),
		Long:  longDesc,
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			urlPath := args[0]
			viewIndex := -1
			itemIndex := -1

			argIdx := 1
			if depth >= 2 {
				var err error
				viewIndex, err = strconv.Atoi(args[argIdx])
				if err != nil {
					return fmt.Errorf("invalid view index: %s", args[argIdx])
				}
				argIdx++
			}

			var err error
			itemIndex, err = strconv.Atoi(args[argIdx])
			if err != nil {
				return fmt.Errorf("invalid %s index: %s", name, args[argIdx])
			}

			textMode := getTextMode()
			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			config, err := fetchDashboardConfig(ws, urlPath)
			if err != nil {
				return err
			}
			if config == nil {
				return fmt.Errorf("invalid dashboard config")
			}

			secIdx := sectionFlag
			if cfg.HasSectionFlag {
				secIdx, _ = cmd.Flags().GetInt("section")
			}

			nav, err := navigateToResourceArray(config, cfg.PathFromConfig, viewIndex, secIdx)
			if err != nil {
				return err
			}

			if nav.Items == nil {
				return fmt.Errorf("no %ss found", name)
			}

			if itemIndex < 0 || itemIndex >= len(nav.Items) {
				return fmt.Errorf("%s index %d out of range (0-%d)", name, itemIndex, len(nav.Items)-1)
			}

			item := nav.Items[itemIndex]
			data := mapItemWithIndex(item, itemIndex, cfg.ItemCanBeString)

			output.PrintOutput(data, textMode, "")
			return nil
		},
	}

	if cfg.HasSectionFlag {
		getCmd.Flags().IntVarP(&sectionFlag, "section", "s", -1, "Section index (if "+name+" is in a section)")
	}

	parentCmd.AddCommand(getCmd)
}

// registerDashResourceGetWithFlags handles view/section/card get commands that
// accept --dashboard / --view / --index flags with positional fallbacks.
func registerDashResourceGetWithFlags(parentCmd *cobra.Command, cfg DashboardResourceConfig, depth int) {
	name := cfg.ResourceName
	sectionFlag := -1

	// Build use string with optional positional args
	useParts := []string{"[dashboard_url_path]"}
	if depth >= 2 {
		useParts = append(useParts, "[view_index]")
	}
	useParts = append(useParts, fmt.Sprintf("[%s_index]", name))
	getUse := "get " + strings.Join(useParts, " ")

	maxArgs := len(useParts)
	if depth == 3 {
		maxArgs = 3 // urlPath, viewIndex, cardIndex (section via flag)
	}

	longDesc := fmt.Sprintf("Get a specific %s from a %s by index.", name, listParentNoun(depth))
	if depth == 3 {
		longDesc = fmt.Sprintf("Get a specific %s from a section by index.\n\nIf section is not specified, uses the last section.", name)
	}

	// Closure-local flag variables
	var dashboardFlag string
	var viewFlag int = -1
	var indexFlag int = -1

	getCmd := &cobra.Command{
		Use:   getUse,
		Short: fmt.Sprintf("Get a specific %s", name),
		Long:  longDesc,
		Args:  cobra.MaximumNArgs(maxArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve dashboard URL path
			urlPath := dashboardFlag
			if urlPath == "" && len(args) > 0 {
				urlPath = args[0]
			}
			if urlPath == "" {
				return fmt.Errorf("dashboard URL path is required (use --dashboard flag or first positional argument)")
			}

			// Resolve view index (for depth >= 2)
			viewIndex := -1
			argIdx := 1
			if depth >= 2 {
				viewIndex = viewFlag
				if viewIndex < 0 && len(args) > argIdx {
					var err error
					viewIndex, err = strconv.Atoi(args[argIdx])
					if err != nil {
						return fmt.Errorf("invalid view index: %s", args[argIdx])
					}
				}
				if viewIndex < 0 {
					return fmt.Errorf("view index is required (use --view flag or %s positional argument)", ordinal(argIdx+1))
				}
				argIdx++
			}

			// Resolve item index
			itemIndex := indexFlag
			if itemIndex < 0 && len(args) > argIdx {
				var err error
				itemIndex, err = strconv.Atoi(args[argIdx])
				if err != nil {
					return fmt.Errorf("invalid %s index: %s", name, args[argIdx])
				}
			}
			if itemIndex < 0 {
				return fmt.Errorf("%s index is required (use --index flag or %s positional argument)", name, ordinal(argIdx+1))
			}

			textMode := getTextMode()
			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			config, err := fetchDashboardConfig(ws, urlPath)
			if err != nil {
				return err
			}
			if config == nil {
				return fmt.Errorf("invalid dashboard config")
			}

			secIdx := sectionFlag
			if cfg.HasSectionFlag {
				secIdx, _ = cmd.Flags().GetInt("section")
			}

			nav, err := navigateToResourceArray(config, cfg.PathFromConfig, viewIndex, secIdx)
			if err != nil {
				return err
			}

			if nav.Items == nil {
				return fmt.Errorf("no %ss in %s", name, listParentNoun(depth))
			}

			if itemIndex < 0 || itemIndex >= len(nav.Items) {
				return fmt.Errorf("%s index %d out of range (0-%d)", name, itemIndex, len(nav.Items)-1)
			}

			item := nav.Items[itemIndex]
			// For non-badge items, set index on original map
			if !cfg.ItemCanBeString {
				if m, ok := item.(map[string]interface{}); ok {
					m["index"] = itemIndex
				}
				output.PrintOutput(item, textMode, "")
			} else {
				data := mapItemWithIndex(item, itemIndex, true)
				output.PrintOutput(data, textMode, "")
			}
			return nil
		},
	}

	getCmd.Flags().StringVar(&dashboardFlag, "dashboard", "", "Dashboard URL path")
	if depth >= 2 {
		getCmd.Flags().IntVar(&viewFlag, "view", -1, "View index")
	}
	getCmd.Flags().IntVar(&indexFlag, "index", -1, capitalize(name)+" index")
	if cfg.HasSectionFlag {
		getCmd.Flags().IntVarP(&sectionFlag, "section", "s", -1, "Section index (if "+name+" is in a section)")
	}

	parentCmd.AddCommand(getCmd)
}

func ordinal(n int) string {
	switch n {
	case 1:
		return "first"
	case 2:
		return "second"
	case 3:
		return "third"
	default:
		return fmt.Sprintf("%d", n)
	}
}

// ──────────────────────────────────────────────────────────
// CREATE
// ──────────────────────────────────────────────────────────

func registerDashResourceCreate(parentCmd *cobra.Command, cfg DashboardResourceConfig, depth int) {
	name := cfg.ResourceName
	sectionFlag := -1

	// Closure-local data/file/format flags
	var dataFlag, fileFlag, formatFlag string

	// Closure-local custom flag values (we store as map for flexibility)
	customFlagValues := make(map[string]*string)
	for _, f := range cfg.CreateFlags {
		val := f.Default
		customFlagValues[f.Name] = &val
	}

	// Args: depth determines positional count
	// view: ExactArgs(1) [urlPath]
	// badge/section: ExactArgs(2) [urlPath, viewIndex]
	// card: RangeArgs(1,2) if CreateViewArgOptional, else ExactArgs(2)
	argCount := depth
	if depth == 3 {
		argCount = 2
	}
	if depth == 1 {
		argCount = 1
	}

	createUse := fmt.Sprintf("create <dashboard_url_path>")
	if depth >= 2 && !cfg.CreateViewArgOptional {
		createUse = fmt.Sprintf("create <dashboard_url_path> <view_index>")
	} else if cfg.CreateViewArgOptional {
		createUse = fmt.Sprintf("create <dashboard_url_path> [view_index]")
	}

	longDesc := fmt.Sprintf("Create a new %s in a dashboard.", name)
	if cfg.CreateLongDesc != "" {
		longDesc = cfg.CreateLongDesc
	}

	var argsValidator cobra.PositionalArgs
	if cfg.CreateViewArgOptional {
		argsValidator = cobra.RangeArgs(1, 2)
	} else {
		argsValidator = cobra.ExactArgs(argCount)
	}

	createCmd := &cobra.Command{
		Use:   createUse,
		Short: fmt.Sprintf("Create a new %s", name),
		Long:  longDesc,
		Args:  argsValidator,
		RunE: func(cmd *cobra.Command, args []string) error {
			urlPath := args[0]
			viewIndex := -1
			if depth >= 2 && len(args) > 1 {
				var err error
				viewIndex, err = strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid view index: %s", args[1])
				}
			}

			textMode := getTextMode()

			// ── Handle badge-specific entity/type shorthand ──
			if cfg.ItemCanBeString {
				return runBadgeStyleCreate(cmd, cfg, urlPath, viewIndex, textMode, dataFlag, fileFlag, formatFlag, customFlagValues)
			}

			// ── Standard map-based create ──
			var resourceConfig map[string]interface{}
			if dataFlag != "" || fileFlag != "" {
				var err error
				resourceConfig, err = input.ParseInput(dataFlag, fileFlag, formatFlag)
				if err != nil {
					return err
				}
			} else {
				resourceConfig = make(map[string]interface{})
			}

			// Apply custom flags
			for _, f := range cfg.CreateFlags {
				val := *customFlagValues[f.Name]
				if val != "" && val != f.Default {
					resourceConfig[f.ConfigKey] = val
				}
			}

			// Apply defaults
			for k, v := range cfg.CreateDefaults {
				if _, exists := resourceConfig[k]; !exists {
					resourceConfig[k] = v
				}
			}

			// Check required fields
			for _, field := range cfg.CreateRequiredFields {
				if _, ok := resourceConfig[field]; !ok {
					return fmt.Errorf("%s %s is required (use --%s or provide in data)", name, field, field)
				}
			}

			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			config, err := fetchDashboardConfig(ws, urlPath)
			if err != nil {
				return err
			}

			// Auto-scaffold for cards
			if cfg.CreateAutoScaffold {
				return runCardStyleCreate(ws, urlPath, config, resourceConfig, viewIndex, sectionFlag, textMode, cmd, cfg)
			}

			// Standard create path
			if config == nil {
				if depth == 1 {
					config = map[string]interface{}{"views": []interface{}{}}
				} else {
					return fmt.Errorf("invalid dashboard config")
				}
			}

			if depth == 1 {
				// View create
				views, ok := config["views"].([]interface{})
				if !ok {
					views = []interface{}{}
				}
				views = append(views, resourceConfig)
				config["views"] = views

				if err := saveDashboardConfig(ws, urlPath, config); err != nil {
					return err
				}

				resourceConfig["index"] = len(views) - 1
				output.PrintSuccess(resourceConfig, textMode, fmt.Sprintf("%s '%v' created at index %d.", capitalize(name), resourceConfig["title"], len(views)-1))
				return nil
			}

			// Depth 2: badge or section within a view
			secIdx := sectionFlag
			if cfg.HasSectionFlag {
				secIdx, _ = cmd.Flags().GetInt("section")
			}
			nav, err := navigateToResourceArray(config, cfg.PathFromConfig, viewIndex, secIdx)
			if err != nil {
				return err
			}

			items := nav.Items
			if items == nil {
				items = []interface{}{}
			}
			items = append(items, resourceConfig)
			nav.Items = items

			if err := saveBackToConfig(ws, urlPath, nav); err != nil {
				return err
			}

			resourceConfig["index"] = len(items) - 1
			output.PrintSuccess(resourceConfig, textMode, fmt.Sprintf("%s created at index %d.", capitalize(name), len(items)-1))
			return nil
		},
	}

	createCmd.Flags().StringVarP(&dataFlag, "data", "d", "", capitalize(name)+" configuration as JSON")
	createCmd.Flags().StringVarP(&fileFlag, "file", "f", "", "Path to config file")
	createCmd.Flags().StringVar(&formatFlag, "format", "", "Input format (json, yaml)")

	for _, f := range cfg.CreateFlags {
		ptr := customFlagValues[f.Name]
		if f.Short != "" {
			createCmd.Flags().StringVarP(ptr, f.Name, f.Short, f.Default, f.Usage)
		} else {
			createCmd.Flags().StringVar(ptr, f.Name, f.Default, f.Usage)
		}
	}

	if cfg.HasSectionFlag {
		createCmd.Flags().IntVarP(&sectionFlag, "section", "s", -1, "Section index (if "+name+" should be in a section)")
	}

	parentCmd.AddCommand(createCmd)
}

// runBadgeStyleCreate handles badge create which supports entity_id strings.
func runBadgeStyleCreate(cmd *cobra.Command, cfg DashboardResourceConfig, urlPath string, viewIndex int, textMode bool, dataFlag, fileFlag, formatFlag string, customFlagValues map[string]*string) error {
	name := cfg.ResourceName

	var badgeConfig interface{}
	entityVal := ""
	typeVal := ""
	if ptr, ok := customFlagValues["entity"]; ok {
		entityVal = *ptr
	}
	if ptr, ok := customFlagValues["type"]; ok {
		typeVal = *ptr
	}

	if dataFlag != "" || fileFlag != "" {
		parsed, err := input.ParseInput(dataFlag, fileFlag, formatFlag)
		if err != nil {
			return err
		}
		badgeConfig = parsed
	} else if entityVal != "" {
		if typeVal != "" {
			badgeConfig = map[string]interface{}{
				"type":   typeVal,
				"entity": entityVal,
			}
		} else {
			badgeConfig = entityVal
		}
	} else {
		return fmt.Errorf("%s configuration required (use --data, --file, or --entity)", name)
	}

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	config, err := fetchDashboardConfig(ws, urlPath)
	if err != nil {
		return err
	}
	if config == nil {
		return fmt.Errorf("invalid dashboard config")
	}

	nav, err := navigateToResourceArray(config, cfg.PathFromConfig, viewIndex, -1)
	if err != nil {
		return err
	}

	items := nav.Items
	if items == nil {
		items = []interface{}{}
	}
	items = append(items, badgeConfig)
	nav.Items = items

	if err := saveBackToConfig(ws, urlPath, nav); err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  len(items) - 1,
		"config": badgeConfig,
	}
	output.PrintSuccess(resultData, textMode, fmt.Sprintf("%s created at index %d.", capitalize(name), len(items)-1))
	return nil
}

// runCardStyleCreate handles card create with auto-scaffolding of views/sections.
func runCardStyleCreate(ws interface{ SendCommand(string, map[string]interface{}) (interface{}, error) }, urlPath string, config map[string]interface{}, cardConfig map[string]interface{}, viewIndex, sectionFlagVal int, textMode bool, cmd *cobra.Command, cfg DashboardResourceConfig) error {
	if config == nil {
		config = map[string]interface{}{
			"views": []interface{}{},
		}
	}

	views, ok := config["views"].([]interface{})
	if !ok {
		views = []interface{}{}
	}

	// If no views exist, create one
	viewCreated := false
	if len(views) == 0 {
		newView := map[string]interface{}{
			"title":    "Home",
			"sections": []interface{}{},
		}
		views = append(views, newView)
		viewCreated = true
	}

	// Default to last view if not specified
	if viewIndex < 0 {
		viewIndex = len(views) - 1
	}

	if viewIndex >= len(views) {
		return fmt.Errorf("view index %d out of range (0-%d)", viewIndex, len(views)-1)
	}

	view, ok := views[viewIndex].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid view at index %d", viewIndex)
	}

	// Get or create sections
	sections, _ := view["sections"].([]interface{})
	if sections == nil {
		sections = []interface{}{}
	}

	// If no sections exist, create one
	sectionCreated := false
	if len(sections) == 0 {
		newSection := map[string]interface{}{
			"type":  "grid",
			"cards": []interface{}{},
		}
		sections = append(sections, newSection)
		view["sections"] = sections
		sectionCreated = true
	}

	// Determine section index
	sectionIndex := sectionFlagVal
	if cfg.HasSectionFlag {
		secIdx, _ := cmd.Flags().GetInt("section")
		sectionIndex = secIdx
	}
	if sectionIndex < 0 {
		sectionIndex = len(sections) - 1
	}

	if sectionIndex >= len(sections) {
		return fmt.Errorf("section index %d out of range (0-%d)", sectionIndex, len(sections)-1)
	}

	section, ok := sections[sectionIndex].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid section at index %d", sectionIndex)
	}

	cards, _ := section["cards"].([]interface{})
	if cards == nil {
		cards = []interface{}{}
	}
	cards = append(cards, cardConfig)
	newCardIndex := len(cards) - 1
	section["cards"] = cards
	sections[sectionIndex] = section
	view["sections"] = sections

	views[viewIndex] = view
	config["views"] = views

	if err := saveDashboardConfig(ws, urlPath, config); err != nil {
		return err
	}

	cardConfig["index"] = newCardIndex

	// Build descriptive message
	var msgParts []string
	if viewCreated {
		msgParts = append(msgParts, fmt.Sprintf("view %d created", viewIndex))
	}
	if sectionCreated {
		msgParts = append(msgParts, fmt.Sprintf("section %d created", sectionIndex))
	}
	msgParts = append(msgParts, fmt.Sprintf("card created at index %d in view %d section %d", newCardIndex, viewIndex, sectionIndex))

	var msg string
	if len(msgParts) > 1 {
		msg = ""
		for i, part := range msgParts {
			if i == 0 {
				msg = strings.ToUpper(part[:1]) + part[1:]
			} else {
				msg += ", " + part
			}
		}
		msg += "."
	} else {
		msg = "Card created at index " + strconv.Itoa(newCardIndex) + " in view " + strconv.Itoa(viewIndex) + " section " + strconv.Itoa(sectionIndex) + "."
	}

	output.PrintSuccess(cardConfig, textMode, msg)
	return nil
}

// ──────────────────────────────────────────────────────────
// UPDATE
// ──────────────────────────────────────────────────────────

func registerDashResourceUpdate(parentCmd *cobra.Command, cfg DashboardResourceConfig, depth int) {
	name := cfg.ResourceName
	sectionFlag := -1

	var dataFlag, fileFlag, formatFlag string

	// Custom flag values
	customFlagValues := make(map[string]*string)
	for _, f := range cfg.UpdateFlags {
		val := f.Default
		customFlagValues[f.Name] = &val
	}

	// Args: depth+1 for positional (urlPath + indices + itemIndex)
	// view: 2 (urlPath, viewIndex)
	// badge/section: 3 (urlPath, viewIndex, itemIndex)
	// card: 3 (urlPath, viewIndex, cardIndex; section via flag)
	argCount := depth + 1
	if depth == 3 {
		argCount = 3
	}

	updateUseParts := []string{"<dashboard_url_path>"}
	if depth >= 1 {
		updateUseParts = append(updateUseParts, "<view_index>")
	}
	if depth >= 2 {
		updateUseParts = append(updateUseParts, fmt.Sprintf("<%s_index>", name))
	}
	// For depth 1 (view), there's only urlPath + view_index
	updateUse := "update " + strings.Join(updateUseParts, " ")

	longDesc := fmt.Sprintf("Update a %s by index.", name)
	if cfg.UpdateLongDesc != "" {
		longDesc = cfg.UpdateLongDesc
	}

	updateCmd := &cobra.Command{
		Use:   updateUse,
		Short: fmt.Sprintf("Update a %s", name),
		Long:  longDesc,
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			urlPath := args[0]
			viewIndex := -1
			itemIndex := -1

			// For view update: args[1] is the view index, and that IS the item
			if depth == 1 {
				var err error
				itemIndex, err = strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid %s index: %s", name, args[1])
				}
				viewIndex = itemIndex // view IS the item
			} else {
				var err error
				viewIndex, err = strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid view index: %s", args[1])
				}
				itemIndex, err = strconv.Atoi(args[2])
				if err != nil {
					return fmt.Errorf("invalid %s index: %s", name, args[2])
				}
			}

			textMode := getTextMode()
			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			config, err := fetchDashboardConfig(ws, urlPath)
			if err != nil {
				return err
			}
			if config == nil {
				return fmt.Errorf("invalid dashboard config")
			}

			// ── Badge-specific update ──
			if cfg.ItemCanBeString {
				return runBadgeStyleUpdate(cmd, cfg, ws, urlPath, config, viewIndex, itemIndex, textMode, dataFlag, fileFlag, formatFlag, customFlagValues)
			}

			// ── Standard map-based update ──
			secIdx := sectionFlag
			if cfg.HasSectionFlag {
				secIdx, _ = cmd.Flags().GetInt("section")
			}

			nav, err := navigateToResourceArray(config, cfg.PathFromConfig, viewIndex, secIdx)
			if err != nil {
				return err
			}

			if depth == 1 {
				// For views, the items ARE the views, and itemIndex == viewIndex
				// But navigateToResourceArray for depth 1 returns views as Items
				// The viewIndex for navigation was already set correctly
			}

			if nav.Items == nil {
				return fmt.Errorf("no %ss in %s", name, listParentNoun(depth))
			}

			if itemIndex < 0 || itemIndex >= len(nav.Items) {
				return fmt.Errorf("%s index %d out of range (0-%d)", name, itemIndex, len(nav.Items)-1)
			}

			// Get existing item
			existing, ok := nav.Items[itemIndex].(map[string]interface{})
			if !ok {
				existing = make(map[string]interface{})
			}

			// If data or file provided, replace entirely
			if dataFlag != "" || fileFlag != "" {
				newConfig, err := input.ParseInput(dataFlag, fileFlag, formatFlag)
				if err != nil {
					return err
				}
				existing = newConfig
			}

			// Apply flag updates
			for _, f := range cfg.UpdateFlags {
				if cmd.Flags().Changed(f.Name) {
					existing[f.ConfigKey] = *customFlagValues[f.Name]
				}
			}

			nav.Items[itemIndex] = existing

			if err := saveBackToConfig(ws, urlPath, nav); err != nil {
				return err
			}

			existing["index"] = itemIndex
			output.PrintSuccess(existing, textMode, fmt.Sprintf("%s at index %d updated.", capitalize(name), itemIndex))
			return nil
		},
	}

	updateCmd.Flags().StringVarP(&dataFlag, "data", "d", "", capitalize(name)+" configuration as JSON (replaces entire "+name+")")
	updateCmd.Flags().StringVarP(&fileFlag, "file", "f", "", "Path to config file")
	updateCmd.Flags().StringVar(&formatFlag, "format", "", "Input format (json, yaml)")

	for _, f := range cfg.UpdateFlags {
		ptr := customFlagValues[f.Name]
		if f.Short != "" {
			updateCmd.Flags().StringVarP(ptr, f.Name, f.Short, f.Default, f.Usage)
		} else {
			updateCmd.Flags().StringVar(ptr, f.Name, f.Default, f.Usage)
		}
	}

	if cfg.HasSectionFlag {
		updateCmd.Flags().IntVarP(&sectionFlag, "section", "s", -1, "Section index (if "+name+" is in a section)")
	}

	parentCmd.AddCommand(updateCmd)
}

// runBadgeStyleUpdate handles badge update which supports entity_id strings.
func runBadgeStyleUpdate(cmd *cobra.Command, cfg DashboardResourceConfig, ws interface{ SendCommand(string, map[string]interface{}) (interface{}, error) }, urlPath string, config map[string]interface{}, viewIndex, itemIndex int, textMode bool, dataFlag, fileFlag, formatFlag string, customFlagValues map[string]*string) error {
	name := cfg.ResourceName

	nav, err := navigateToResourceArray(config, cfg.PathFromConfig, viewIndex, -1)
	if err != nil {
		return err
	}

	if nav.Items == nil {
		return fmt.Errorf("no %ss in view", name)
	}

	if itemIndex < 0 || itemIndex >= len(nav.Items) {
		return fmt.Errorf("%s index %d out of range (0-%d)", name, itemIndex, len(nav.Items)-1)
	}

	var newItem interface{}

	if dataFlag != "" || fileFlag != "" {
		parsed, err := input.ParseInput(dataFlag, fileFlag, formatFlag)
		if err != nil {
			return err
		}
		newItem = parsed
	} else if cmd.Flags().Changed("entity") {
		entityVal := ""
		if ptr, ok := customFlagValues["entity"]; ok {
			entityVal = *ptr
		}
		newItem = entityVal
	} else {
		return fmt.Errorf("update data required (use --data, --file, or --entity)")
	}

	nav.Items[itemIndex] = newItem

	if err := saveBackToConfig(ws, urlPath, nav); err != nil {
		return err
	}

	resultData := map[string]interface{}{
		"index":  itemIndex,
		"config": newItem,
	}
	output.PrintSuccess(resultData, textMode, fmt.Sprintf("%s at index %d updated.", capitalize(name), itemIndex))
	return nil
}

// ──────────────────────────────────────────────────────────
// DELETE
// ──────────────────────────────────────────────────────────

func registerDashResourceDelete(parentCmd *cobra.Command, cfg DashboardResourceConfig, depth int) {
	name := cfg.ResourceName
	sectionFlag := -1
	var forceFlag bool

	// Same arg count logic as update
	argCount := depth + 1
	if depth == 3 {
		argCount = 3
	}

	deleteUseParts := []string{"<dashboard_url_path>"}
	if depth >= 1 {
		deleteUseParts = append(deleteUseParts, "<view_index>")
	}
	if depth >= 2 {
		deleteUseParts = append(deleteUseParts, fmt.Sprintf("<%s_index>", name))
	}
	deleteUse := "delete " + strings.Join(deleteUseParts, " ")

	longDesc := fmt.Sprintf("Delete a %s by index.", name)
	if cfg.DeleteLongDesc != "" {
		longDesc = cfg.DeleteLongDesc
	}
	if depth == 3 {
		longDesc = fmt.Sprintf("Delete a %s from a section by index.\n\nIf section is not specified, uses the last section.", name)
	}

	deleteCmd := &cobra.Command{
		Use:   deleteUse,
		Short: fmt.Sprintf("Delete a %s", name),
		Long:  longDesc,
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			urlPath := args[0]
			viewIndex := -1
			itemIndex := -1

			if depth == 1 {
				var err error
				itemIndex, err = strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid %s index: %s", name, args[1])
				}
				viewIndex = itemIndex
			} else {
				var err error
				viewIndex, err = strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid view index: %s", args[1])
				}
				itemIndex, err = strconv.Atoi(args[2])
				if err != nil {
					return fmt.Errorf("invalid %s index: %s", name, args[2])
				}
			}

			textMode := getTextMode()
			ws, err := getWSClient()
			if err != nil {
				return err
			}
			defer ws.Close()

			config, err := fetchDashboardConfig(ws, urlPath)
			if err != nil {
				return err
			}
			if config == nil {
				return fmt.Errorf("invalid dashboard config")
			}

			secIdx := sectionFlag
			if cfg.HasSectionFlag {
				secIdx, _ = cmd.Flags().GetInt("section")
			}

			nav, err := navigateToResourceArray(config, cfg.PathFromConfig, viewIndex, secIdx)
			if err != nil {
				return err
			}

			if nav.Items == nil {
				return fmt.Errorf("no %ss found", name)
			}

			if itemIndex < 0 || itemIndex >= len(nav.Items) {
				return fmt.Errorf("%s index %d out of range (0-%d)", name, itemIndex, len(nav.Items)-1)
			}

			// Build confirmation description
			desc := fmt.Sprintf("%s at index %d", name, itemIndex)
			item := nav.Items[itemIndex]
			if m, ok := item.(map[string]interface{}); ok {
				if title, ok := m["title"].(string); ok && title != "" {
					desc = fmt.Sprintf("%s '%s' (index %d)", name, title, itemIndex)
				} else if cardType, ok := m["type"].(string); ok && name == "card" {
					desc = fmt.Sprintf("%s card (index %d)", cardType, itemIndex)
				}
			}

			if err := confirmDelete(forceFlag, textMode, desc); err != nil {
				return err
			}

			// Remove the item
			nav.Items = append(nav.Items[:itemIndex], nav.Items[itemIndex+1:]...)

			if err := saveBackToConfig(ws, urlPath, nav); err != nil {
				return err
			}

			output.PrintSuccess(nil, textMode, fmt.Sprintf("%s at index %d deleted.", capitalize(name), itemIndex))
			return nil
		},
	}

	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation prompt")
	if cfg.HasSectionFlag {
		deleteCmd.Flags().IntVarP(&sectionFlag, "section", "s", -1, "Section index (if "+name+" is in a section)")
	}

	parentCmd.AddCommand(deleteCmd)
}

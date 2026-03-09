package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

// ──────────────────────────────────────────────────────────
// CREATE
// ──────────────────────────────────────────────────────────

func registerDashResourceCreate(parentCmd *cobra.Command, cfg DashboardResourceConfig, depth int) {
	name := cfg.ResourceName
	sectionFlag := -1

	// Closure-local input flags
	var inputFlags InputFlags

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
		Use:     createUse,
		Short:   fmt.Sprintf("Create a new %s", name),
		Long:    longDesc,
		Example: cfg.CreateExample,
		Args:    argsValidator,
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
				return runBadgeStyleCreate(cmd, cfg, urlPath, viewIndex, textMode, &inputFlags, customFlagValues)
			}

			// ── Standard map-based create ──
			var resourceConfig map[string]interface{}
			if inputFlags.Data != "" || inputFlags.File != "" {
				var err error
				resourceConfig, err = inputFlags.Parse()
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

	inputFlags.Register(createCmd)

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
func runBadgeStyleCreate(cmd *cobra.Command, cfg DashboardResourceConfig, urlPath string, viewIndex int, textMode bool, flags *InputFlags, customFlagValues map[string]*string) error {
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

	if flags.Data != "" || flags.File != "" {
		parsed, err := flags.Parse()
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

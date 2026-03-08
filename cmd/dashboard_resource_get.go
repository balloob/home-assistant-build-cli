package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

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

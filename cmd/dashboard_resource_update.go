package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

// ──────────────────────────────────────────────────────────
// UPDATE
// ──────────────────────────────────────────────────────────

func registerDashResourceUpdate(parentCmd *cobra.Command, cfg DashboardResourceConfig, depth int) {
	name := cfg.ResourceName
	sectionFlag := -1

	var inputFlags InputFlags

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
				return runBadgeStyleUpdate(cmd, cfg, ws, urlPath, config, viewIndex, itemIndex, textMode, &inputFlags, customFlagValues)
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
			if inputFlags.Data != "" || inputFlags.File != "" {
				newConfig, err := inputFlags.Parse()
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

	inputFlags.Register(updateCmd)

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
func runBadgeStyleUpdate(cmd *cobra.Command, cfg DashboardResourceConfig, ws interface{ SendCommand(string, map[string]interface{}) (interface{}, error) }, urlPath string, config map[string]interface{}, viewIndex, itemIndex int, textMode bool, flags *InputFlags, customFlagValues map[string]*string) error {
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

	if flags.Data != "" || flags.File != "" {
		parsed, err := flags.Parse()
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

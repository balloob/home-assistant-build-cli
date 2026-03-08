package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

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

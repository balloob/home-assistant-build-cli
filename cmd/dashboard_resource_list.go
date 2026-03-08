package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

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

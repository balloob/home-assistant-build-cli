package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var dashboardListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all dashboards",
	Long:    `List all dashboards in Home Assistant.`,
	GroupID: dashboardGroupCommands,
	RunE:    runDashboardList,
}

var dashboardListFlags *ListFlags

func init() {
	dashboardCmd.AddCommand(dashboardListCmd)
	dashboardListFlags = RegisterListFlags(dashboardListCmd, "url_path")
}

func runDashboardList(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	result, err := ws.SendCommand("lovelace/dashboards/list", nil)
	if err != nil {
		return err
	}

	// Convert to slice for processing
	dashboards, ok := result.([]interface{})
	if !ok {
		output.PrintOutput(result, textMode, "")
		return nil
	}

	if dashboardListFlags.RenderCount(len(dashboards), textMode) {
		return nil
	}
	dashboards = dashboardListFlags.ApplyLimit(dashboards)

	// Handle brief mode — custom text rendering for title=="" fallback
	if dashboardListFlags.Brief {
		if textMode {
			for _, d := range dashboards {
				if dashboard, ok := d.(map[string]interface{}); ok {
					title := getStr(dashboard, "title")
					urlPath := getStr(dashboard, "url_path")
					if title != "" {
						fmt.Printf("%s (%s)\n", title, urlPath)
					} else {
						fmt.Println(urlPath)
					}
				}
			}
		} else {
			var brief []map[string]interface{}
			for _, d := range dashboards {
				if dashboard, ok := d.(map[string]interface{}); ok {
					brief = append(brief, map[string]interface{}{
						"url_path": dashboard["url_path"],
						"title":    dashboard["title"],
					})
				}
			}
			output.PrintOutput(brief, false, "")
		}
		return nil
	}

	// Full output
	if textMode {
		if len(dashboards) == 0 {
			fmt.Println("No dashboards.")
			return nil
		}
		for _, d := range dashboards {
			if dashboard, ok := d.(map[string]interface{}); ok {
				printDashboardText(dashboard)
				fmt.Println()
			}
		}
	} else {
		output.PrintOutput(dashboards, false, "")
	}
	return nil
}

func printDashboardText(d map[string]interface{}) {
	title := getStr(d, "title")
	urlPath := getStr(d, "url_path")

	if title != "" {
		fmt.Printf("%s:\n", title)
	} else {
		fmt.Printf("%s:\n", urlPath)
	}

	if urlPath != "" {
		fmt.Printf("  path: %s\n", urlPath)
	}
	if mode := getStr(d, "mode"); mode != "" {
		fmt.Printf("  mode: %s\n", mode)
	}
	if requireAdmin, ok := d["require_admin"].(bool); ok && requireAdmin {
		fmt.Println("  require_admin: yes")
	}
	if showInSidebar, ok := d["show_in_sidebar"].(bool); ok && !showInSidebar {
		fmt.Println("  show_in_sidebar: no")
	}
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func formatBoolYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func joinNonEmpty(sep string, parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, sep)
}

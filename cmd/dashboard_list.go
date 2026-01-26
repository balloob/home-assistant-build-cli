package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dashboardListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all dashboards",
	Long:    `List all dashboards in Home Assistant.`,
	GroupID: dashboardGroupCommands,
	RunE:    runDashboardList,
}

var (
	dashboardListCount bool
	dashboardListBrief bool
	dashboardListLimit int
)

func init() {
	dashboardCmd.AddCommand(dashboardListCmd)
	dashboardListCmd.Flags().BoolVarP(&dashboardListCount, "count", "c", false, "Return only the count of items")
	dashboardListCmd.Flags().BoolVarP(&dashboardListBrief, "brief", "b", false, "Return minimal fields (url_path and title only)")
	dashboardListCmd.Flags().IntVarP(&dashboardListLimit, "limit", "n", 0, "Limit results to N items")
}

func runDashboardList(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
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
		client.PrintOutput(result, textMode, "")
		return nil
	}

	// Handle count mode
	if dashboardListCount {
		if textMode {
			fmt.Printf("Count: %d\n", len(dashboards))
		} else {
			client.PrintOutput(map[string]interface{}{"count": len(dashboards)}, false, "")
		}
		return nil
	}

	// Apply limit
	if dashboardListLimit > 0 && len(dashboards) > dashboardListLimit {
		dashboards = dashboards[:dashboardListLimit]
	}

	// Handle brief mode
	if dashboardListBrief {
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
			client.PrintOutput(brief, false, "")
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
		client.PrintOutput(dashboards, false, "")
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

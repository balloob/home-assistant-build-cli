package cmd

import (
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
		client.PrintOutput(map[string]interface{}{"count": len(dashboards)}, textMode, "")
		return nil
	}

	// Apply limit
	if dashboardListLimit > 0 && len(dashboards) > dashboardListLimit {
		dashboards = dashboards[:dashboardListLimit]
	}

	// Handle brief mode
	if dashboardListBrief {
		var brief []map[string]interface{}
		for _, d := range dashboards {
			if dashboard, ok := d.(map[string]interface{}); ok {
				brief = append(brief, map[string]interface{}{
					"url_path": dashboard["url_path"],
					"title":    dashboard["title"],
				})
			}
		}
		client.PrintOutput(brief, textMode, "")
		return nil
	}

	client.PrintOutput(dashboards, textMode, "")
	return nil
}

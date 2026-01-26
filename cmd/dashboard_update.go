package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dashboardUpdateTitle         string
	dashboardUpdateIcon          string
	dashboardUpdateShowInSidebar *bool
	dashboardUpdateRequireAdmin  *bool
)

var dashboardUpdateCmd = &cobra.Command{
	Use:     "update <dashboard_id>",
	Short:   "Update a dashboard",
	Long:    `Update dashboard settings like title, icon, sidebar visibility, and admin requirement.`,
	GroupID: dashboardGroupCommands,
	Args:    cobra.ExactArgs(1),
	RunE:    runDashboardUpdate,
}

func init() {
	dashboardCmd.AddCommand(dashboardUpdateCmd)
	dashboardUpdateCmd.Flags().StringVar(&dashboardUpdateTitle, "title", "", "Dashboard title")
	dashboardUpdateCmd.Flags().StringVar(&dashboardUpdateIcon, "icon", "", "Dashboard icon (e.g., mdi:home)")

	// Use custom bool flags to detect if they were set
	var sidebar, admin bool
	dashboardUpdateCmd.Flags().BoolVar(&sidebar, "sidebar", true, "Show in sidebar")
	dashboardUpdateCmd.Flags().BoolVar(&admin, "require-admin", false, "Require admin access")
}

func runDashboardUpdate(cmd *cobra.Command, args []string) error {
	dashboardID := args[0]
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

	params := map[string]interface{}{
		"dashboard_id": dashboardID,
	}

	if cmd.Flags().Changed("title") {
		params["title"] = dashboardUpdateTitle
	}
	if cmd.Flags().Changed("icon") {
		params["icon"] = dashboardUpdateIcon
	}
	if cmd.Flags().Changed("sidebar") {
		sidebar, _ := cmd.Flags().GetBool("sidebar")
		params["show_in_sidebar"] = sidebar
	}
	if cmd.Flags().Changed("require-admin") {
		admin, _ := cmd.Flags().GetBool("require-admin")
		params["require_admin"] = admin
	}

	result, err := ws.SendCommand("lovelace/dashboards/update", params)
	if err != nil {
		return err
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Dashboard %s updated successfully.", dashboardID))
	return nil
}

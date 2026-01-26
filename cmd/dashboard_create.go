package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dashboardCreateTitle         string
	dashboardCreateIcon          string
	dashboardCreateShowInSidebar bool
	dashboardCreateRequireAdmin  bool
	dashboardCreateUrlPath       string
)

var dashboardCreateCmd = &cobra.Command{
	Use:   "create [url_path]",
	Short: "Create a new dashboard",
	Long: `Create a new storage-mode dashboard.

The dashboard is initialized with a single section-based view, ready for adding cards.`,
	GroupID: dashboardGroupCommands,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runDashboardCreate,
}

func init() {
	dashboardCmd.AddCommand(dashboardCreateCmd)
	dashboardCreateCmd.Flags().StringVar(&dashboardCreateTitle, "title", "", "Dashboard title (required)")
	dashboardCreateCmd.Flags().StringVar(&dashboardCreateUrlPath, "url-path", "", "Dashboard URL path (must contain a hyphen)")
	dashboardCreateCmd.Flags().StringVar(&dashboardCreateIcon, "icon", "", "Dashboard icon (e.g., mdi:home)")
	dashboardCreateCmd.Flags().BoolVar(&dashboardCreateShowInSidebar, "sidebar", true, "Show in sidebar")
	dashboardCreateCmd.Flags().BoolVar(&dashboardCreateRequireAdmin, "require-admin", false, "Require admin access")
	dashboardCreateCmd.MarkFlagRequired("title")
}

func runDashboardCreate(cmd *cobra.Command, args []string) error {
	// Determine url_path from flag or positional argument
	var urlPath string
	if dashboardCreateUrlPath != "" {
		urlPath = dashboardCreateUrlPath
	} else if len(args) > 0 {
		urlPath = args[0]
	} else {
		return fmt.Errorf("url_path is required (provide as argument or via --url-path flag)")
	}

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
		"url_path":        urlPath,
		"mode":            "storage",
		"title":           dashboardCreateTitle,
		"show_in_sidebar": dashboardCreateShowInSidebar,
		"require_admin":   dashboardCreateRequireAdmin,
	}
	if dashboardCreateIcon != "" {
		params["icon"] = dashboardCreateIcon
	}

	result, err := ws.SendCommand("lovelace/dashboards/create", params)
	if err != nil {
		return err
	}

	// Initialize the dashboard with a section-based view
	initialConfig := map[string]interface{}{
		"views": []map[string]interface{}{
			{
				"type":     "sections",
				"title":    "Home",
				"path":     "home",
				"sections": []interface{}{},
			},
		},
	}

	saveParams := map[string]interface{}{
		"url_path": urlPath,
		"config":   initialConfig,
	}

	_, err = ws.SendCommand("lovelace/config/save", saveParams)
	if err != nil {
		// Dashboard was created but config failed - warn but don't fail
		client.PrintSuccess(result, textMode, fmt.Sprintf("Dashboard %s created, but initial config failed: %v", urlPath, err))
		return nil
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Dashboard %s created with initial view.", urlPath))
	return nil
}

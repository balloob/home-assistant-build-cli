package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
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
	Example: `hab dashboard create kitchen-dashboard --title "Kitchen"
hab dashboard create --url-path office-dashboard --title "Office" --icon mdi:desk`,
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
	if dashboardCreateUrlPath != "" && len(args) > 0 && dashboardCreateUrlPath != args[0] {
		return fmt.Errorf("conflicting dashboard URL path values: positional %q does not match flag value %q", args[0], dashboardCreateUrlPath)
	}
	if dashboardCreateUrlPath != "" {
		urlPath = dashboardCreateUrlPath
		noteResolution("dashboard_url_path", "flag")
	} else if len(args) > 0 {
		urlPath = args[0]
		noteResolution("dashboard_url_path", "positional")
	} else {
		return fmt.Errorf("url_path is required (provide as argument or via --url-path flag)")
	}

	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
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
		noteFallback("dashboard created successfully, but initial section-based config save failed")
		output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("Dashboard %s created, but initial config failed: %v", urlPath, err), output.EnvelopeContext{Operation: "create", ResourceType: "dashboard"})
		return nil
	}

	output.PrintSuccessWithContext(result, textMode, fmt.Sprintf("Dashboard %s created with initial view.", urlPath), output.EnvelopeContext{Operation: "create", ResourceType: "dashboard"})
	return nil
}

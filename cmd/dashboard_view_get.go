package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	viewGetDashboard string
	viewGetIndex     int
)

var viewGetCmd = &cobra.Command{
	Use:   "get [dashboard_url_path] [view_index]",
	Short: "Get a specific view",
	Long:  `Get a specific view from a dashboard by index.`,
	Args:  cobra.MaximumNArgs(2),
	RunE:  runViewGet,
}

func init() {
	dashboardViewCmd.AddCommand(viewGetCmd)
	viewGetCmd.Flags().StringVar(&viewGetDashboard, "dashboard", "", "Dashboard URL path")
	viewGetCmd.Flags().IntVar(&viewGetIndex, "index", -1, "View index")
}

func runViewGet(cmd *cobra.Command, args []string) error {
	urlPath := viewGetDashboard
	if urlPath == "" && len(args) > 0 {
		urlPath = args[0]
	}
	if urlPath == "" {
		return fmt.Errorf("dashboard URL path is required (use --dashboard flag or first positional argument)")
	}

	viewIndex := viewGetIndex
	if viewIndex < 0 && len(args) > 1 {
		var err error
		viewIndex, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid view index: %s", args[1])
		}
	}
	if viewIndex < 0 {
		return fmt.Errorf("view index is required (use --index flag or second positional argument)")
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

	params := map[string]interface{}{}
	if urlPath != "lovelace" {
		params["url_path"] = urlPath
	}

	result, err := ws.SendCommand("lovelace/config", params)
	if err != nil {
		return err
	}

	config, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid dashboard config")
	}

	views, ok := config["views"].([]interface{})
	if !ok {
		return fmt.Errorf("no views in dashboard")
	}

	if viewIndex < 0 || viewIndex >= len(views) {
		return fmt.Errorf("view index %d out of range (0-%d)", viewIndex, len(views)-1)
	}

	view := views[viewIndex]
	if viewMap, ok := view.(map[string]interface{}); ok {
		viewMap["index"] = viewIndex
	}

	client.PrintOutput(view, textMode, "")
	return nil
}

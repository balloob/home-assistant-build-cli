package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	viewUpdateData   string
	viewUpdateFile   string
	viewUpdateFormat string
	viewUpdateTitle  string
	viewUpdateIcon   string
	viewUpdatePath   string
)

var viewUpdateCmd = &cobra.Command{
	Use:   "update <dashboard_url_path> <view_index>",
	Short: "Update a view",
	Long:  `Update a view in a dashboard by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runViewUpdate,
}

func init() {
	dashboardViewCmd.AddCommand(viewUpdateCmd)
	viewUpdateCmd.Flags().StringVarP(&viewUpdateData, "data", "d", "", "View configuration as JSON (replaces entire view)")
	viewUpdateCmd.Flags().StringVarP(&viewUpdateFile, "file", "f", "", "Path to config file")
	viewUpdateCmd.Flags().StringVar(&viewUpdateFormat, "format", "", "Input format (json, yaml)")
	viewUpdateCmd.Flags().StringVar(&viewUpdateTitle, "title", "", "View title")
	viewUpdateCmd.Flags().StringVar(&viewUpdateIcon, "icon", "", "View icon (e.g., mdi:home)")
	viewUpdateCmd.Flags().StringVar(&viewUpdatePath, "path", "", "View path (URL slug)")
}

func runViewUpdate(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	viewIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid view index: %s", args[1])
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

	// Get current dashboard config
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

	// Get existing view
	existingView, ok := views[viewIndex].(map[string]interface{})
	if !ok {
		existingView = make(map[string]interface{})
	}

	// If data or file provided, replace the entire view config
	if viewUpdateData != "" || viewUpdateFile != "" {
		newConfig, err := input.ParseInput(viewUpdateData, viewUpdateFile, viewUpdateFormat)
		if err != nil {
			return err
		}
		existingView = newConfig
	}

	// Apply flag updates
	if cmd.Flags().Changed("title") {
		existingView["title"] = viewUpdateTitle
	}
	if cmd.Flags().Changed("icon") {
		existingView["icon"] = viewUpdateIcon
	}
	if cmd.Flags().Changed("path") {
		existingView["path"] = viewUpdatePath
	}

	views[viewIndex] = existingView
	config["views"] = views

	// Save the config
	saveParams := map[string]interface{}{
		"config": config,
	}
	if urlPath != "lovelace" {
		saveParams["url_path"] = urlPath
	}

	_, err = ws.SendCommand("lovelace/config/save", saveParams)
	if err != nil {
		return err
	}

	existingView["index"] = viewIndex
	client.PrintSuccess(existingView, textMode, fmt.Sprintf("View at index %d updated.", viewIndex))
	return nil
}

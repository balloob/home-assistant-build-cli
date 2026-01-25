package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	viewCreateData   string
	viewCreateFile   string
	viewCreateFormat string
	viewCreateTitle  string
	viewCreateIcon   string
	viewCreatePath   string
)

var viewCreateCmd = &cobra.Command{
	Use:   "create <dashboard_url_path>",
	Short: "Create a new view",
	Long:  `Create a new view in a dashboard.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runViewCreate,
}

func init() {
	dashboardViewCmd.AddCommand(viewCreateCmd)
	viewCreateCmd.Flags().StringVarP(&viewCreateData, "data", "d", "", "View configuration as JSON")
	viewCreateCmd.Flags().StringVarP(&viewCreateFile, "file", "f", "", "Path to config file")
	viewCreateCmd.Flags().StringVar(&viewCreateFormat, "format", "", "Input format (json, yaml)")
	viewCreateCmd.Flags().StringVar(&viewCreateTitle, "title", "", "View title")
	viewCreateCmd.Flags().StringVar(&viewCreateIcon, "icon", "", "View icon (e.g., mdi:home)")
	viewCreateCmd.Flags().StringVar(&viewCreatePath, "path", "", "View path (URL slug)")
}

func runViewCreate(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	var viewConfig map[string]interface{}

	// If data or file provided, parse it
	if viewCreateData != "" || viewCreateFile != "" {
		var err error
		viewConfig, err = input.ParseInput(viewCreateData, viewCreateFile, viewCreateFormat)
		if err != nil {
			return err
		}
	} else {
		viewConfig = make(map[string]interface{})
	}

	// Apply flags if provided
	if viewCreateTitle != "" {
		viewConfig["title"] = viewCreateTitle
	}
	if viewCreateIcon != "" {
		viewConfig["icon"] = viewCreateIcon
	}
	if viewCreatePath != "" {
		viewConfig["path"] = viewCreatePath
	}

	// Ensure title is set
	if _, ok := viewConfig["title"]; !ok {
		return fmt.Errorf("view title is required (use --title or provide in data)")
	}

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
		config = map[string]interface{}{
			"views": []interface{}{},
		}
	}

	views, ok := config["views"].([]interface{})
	if !ok {
		views = []interface{}{}
	}

	// Add the new view
	views = append(views, viewConfig)
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

	viewConfig["index"] = len(views) - 1
	client.PrintSuccess(viewConfig, textMode, fmt.Sprintf("View '%s' created at index %d.", viewConfig["title"], len(views)-1))
	return nil
}

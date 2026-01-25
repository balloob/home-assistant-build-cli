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
	sectionCreateData   string
	sectionCreateFile   string
	sectionCreateFormat string
	sectionCreateTitle  string
	sectionCreateType   string
)

var sectionCreateCmd = &cobra.Command{
	Use:   "create <dashboard_url_path> <view_index>",
	Short: "Create a new section",
	Long:  `Create a new section in a dashboard view.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runSectionCreate,
}

func init() {
	dashboardSectionCmd.AddCommand(sectionCreateCmd)
	sectionCreateCmd.Flags().StringVarP(&sectionCreateData, "data", "d", "", "Section configuration as JSON")
	sectionCreateCmd.Flags().StringVarP(&sectionCreateFile, "file", "f", "", "Path to config file")
	sectionCreateCmd.Flags().StringVar(&sectionCreateFormat, "format", "", "Input format (json, yaml)")
	sectionCreateCmd.Flags().StringVar(&sectionCreateTitle, "title", "", "Section title")
	sectionCreateCmd.Flags().StringVar(&sectionCreateType, "type", "", "Section type (e.g., grid)")
}

func runSectionCreate(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	viewIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid view index: %s", args[1])
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	var sectionConfig map[string]interface{}

	// If data or file provided, parse it
	if sectionCreateData != "" || sectionCreateFile != "" {
		sectionConfig, err = input.ParseInput(sectionCreateData, sectionCreateFile, sectionCreateFormat)
		if err != nil {
			return err
		}
	} else {
		sectionConfig = make(map[string]interface{})
	}

	// Apply flags if provided
	if sectionCreateTitle != "" {
		sectionConfig["title"] = sectionCreateTitle
	}
	if sectionCreateType != "" {
		sectionConfig["type"] = sectionCreateType
	}

	// Initialize cards array if not present
	if _, ok := sectionConfig["cards"]; !ok {
		sectionConfig["cards"] = []interface{}{}
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
		return fmt.Errorf("invalid dashboard config")
	}

	views, ok := config["views"].([]interface{})
	if !ok {
		return fmt.Errorf("no views in dashboard")
	}

	if viewIndex < 0 || viewIndex >= len(views) {
		return fmt.Errorf("view index %d out of range (0-%d)", viewIndex, len(views)-1)
	}

	view, ok := views[viewIndex].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid view at index %d", viewIndex)
	}

	sections, ok := view["sections"].([]interface{})
	if !ok {
		sections = []interface{}{}
	}

	// Add the new section
	sections = append(sections, sectionConfig)
	view["sections"] = sections
	views[viewIndex] = view
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

	sectionConfig["index"] = len(sections) - 1
	client.PrintSuccess(sectionConfig, textMode, fmt.Sprintf("Section created at index %d.", len(sections)-1))
	return nil
}

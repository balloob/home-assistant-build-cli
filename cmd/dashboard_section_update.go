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
	sectionUpdateData   string
	sectionUpdateFile   string
	sectionUpdateFormat string
	sectionUpdateTitle  string
	sectionUpdateType   string
)

var sectionUpdateCmd = &cobra.Command{
	Use:   "update <dashboard_url_path> <view_index> <section_index>",
	Short: "Update a section",
	Long:  `Update a section in a view by index.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runSectionUpdate,
}

func init() {
	dashboardSectionCmd.AddCommand(sectionUpdateCmd)
	sectionUpdateCmd.Flags().StringVarP(&sectionUpdateData, "data", "d", "", "Section configuration as JSON (replaces entire section)")
	sectionUpdateCmd.Flags().StringVarP(&sectionUpdateFile, "file", "f", "", "Path to config file")
	sectionUpdateCmd.Flags().StringVar(&sectionUpdateFormat, "format", "", "Input format (json, yaml)")
	sectionUpdateCmd.Flags().StringVar(&sectionUpdateTitle, "title", "", "Section title")
	sectionUpdateCmd.Flags().StringVar(&sectionUpdateType, "type", "", "Section type")
}

func runSectionUpdate(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	viewIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid view index: %s", args[1])
	}
	sectionIndex, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid section index: %s", args[2])
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

	view, ok := views[viewIndex].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid view at index %d", viewIndex)
	}

	sections, ok := view["sections"].([]interface{})
	if !ok {
		return fmt.Errorf("no sections in view")
	}

	if sectionIndex < 0 || sectionIndex >= len(sections) {
		return fmt.Errorf("section index %d out of range (0-%d)", sectionIndex, len(sections)-1)
	}

	// Get existing section
	existingSection, ok := sections[sectionIndex].(map[string]interface{})
	if !ok {
		existingSection = make(map[string]interface{})
	}

	// If data or file provided, replace the entire section config
	if sectionUpdateData != "" || sectionUpdateFile != "" {
		newConfig, err := input.ParseInput(sectionUpdateData, sectionUpdateFile, sectionUpdateFormat)
		if err != nil {
			return err
		}
		existingSection = newConfig
	}

	// Apply flag updates
	if cmd.Flags().Changed("title") {
		existingSection["title"] = sectionUpdateTitle
	}
	if cmd.Flags().Changed("type") {
		existingSection["type"] = sectionUpdateType
	}

	sections[sectionIndex] = existingSection
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

	existingSection["index"] = sectionIndex
	client.PrintSuccess(existingSection, textMode, fmt.Sprintf("Section at index %d updated.", sectionIndex))
	return nil
}

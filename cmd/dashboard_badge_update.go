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
	badgeUpdateData   string
	badgeUpdateFile   string
	badgeUpdateFormat string
	badgeUpdateEntity string
)

var badgeUpdateCmd = &cobra.Command{
	Use:   "update <dashboard_url_path> <view_index> <badge_index>",
	Short: "Update a badge",
	Long:  `Update a badge in a view by index.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runBadgeUpdate,
}

func init() {
	dashboardBadgeCmd.AddCommand(badgeUpdateCmd)
	badgeUpdateCmd.Flags().StringVarP(&badgeUpdateData, "data", "d", "", "Badge configuration as JSON (replaces entire badge)")
	badgeUpdateCmd.Flags().StringVarP(&badgeUpdateFile, "file", "f", "", "Path to config file")
	badgeUpdateCmd.Flags().StringVar(&badgeUpdateFormat, "format", "", "Input format (json, yaml)")
	badgeUpdateCmd.Flags().StringVar(&badgeUpdateEntity, "entity", "", "Entity ID for simple badge")
}

func runBadgeUpdate(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	viewIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid view index: %s", args[1])
	}
	badgeIndex, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid badge index: %s", args[2])
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

	badges, ok := view["badges"].([]interface{})
	if !ok {
		return fmt.Errorf("no badges in view")
	}

	if badgeIndex < 0 || badgeIndex >= len(badges) {
		return fmt.Errorf("badge index %d out of range (0-%d)", badgeIndex, len(badges)-1)
	}

	var newBadge interface{}

	// If data or file provided, replace entirely
	if badgeUpdateData != "" || badgeUpdateFile != "" {
		parsed, err := input.ParseInput(badgeUpdateData, badgeUpdateFile, badgeUpdateFormat)
		if err != nil {
			return err
		}
		newBadge = parsed
	} else if cmd.Flags().Changed("entity") {
		// Update to simple entity badge
		newBadge = badgeUpdateEntity
	} else {
		return fmt.Errorf("update data required (use --data, --file, or --entity)")
	}

	badges[badgeIndex] = newBadge
	view["badges"] = badges
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

	resultData := map[string]interface{}{
		"index":  badgeIndex,
		"config": newBadge,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Badge at index %d updated.", badgeIndex))
	return nil
}

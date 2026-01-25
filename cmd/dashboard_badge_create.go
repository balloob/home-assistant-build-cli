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
	badgeCreateData   string
	badgeCreateFile   string
	badgeCreateFormat string
	badgeCreateEntity string
	badgeCreateType   string
)

var badgeCreateCmd = &cobra.Command{
	Use:   "create <dashboard_url_path> <view_index>",
	Short: "Create a new badge",
	Long:  `Create a new badge in a dashboard view.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runBadgeCreate,
}

func init() {
	dashboardBadgeCmd.AddCommand(badgeCreateCmd)
	badgeCreateCmd.Flags().StringVarP(&badgeCreateData, "data", "d", "", "Badge configuration as JSON")
	badgeCreateCmd.Flags().StringVarP(&badgeCreateFile, "file", "f", "", "Path to config file")
	badgeCreateCmd.Flags().StringVar(&badgeCreateFormat, "format", "", "Input format (json, yaml)")
	badgeCreateCmd.Flags().StringVar(&badgeCreateEntity, "entity", "", "Entity ID for simple badge")
	badgeCreateCmd.Flags().StringVar(&badgeCreateType, "type", "", "Badge type (e.g., entity)")
}

func runBadgeCreate(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	viewIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid view index: %s", args[1])
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	var badgeConfig interface{}

	// If data or file provided, parse it
	if badgeCreateData != "" || badgeCreateFile != "" {
		config, err := input.ParseInput(badgeCreateData, badgeCreateFile, badgeCreateFormat)
		if err != nil {
			return err
		}
		badgeConfig = config
	} else if badgeCreateEntity != "" {
		// Simple entity badge
		if badgeCreateType != "" {
			badgeConfig = map[string]interface{}{
				"type":   badgeCreateType,
				"entity": badgeCreateEntity,
			}
		} else {
			badgeConfig = badgeCreateEntity
		}
	} else {
		return fmt.Errorf("badge configuration required (use --data, --file, or --entity)")
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

	badges, ok := view["badges"].([]interface{})
	if !ok {
		badges = []interface{}{}
	}

	// Add the new badge
	badges = append(badges, badgeConfig)
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
		"index":  len(badges) - 1,
		"config": badgeConfig,
	}
	client.PrintSuccess(resultData, textMode, fmt.Sprintf("Badge created at index %d.", len(badges)-1))
	return nil
}

package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var badgeGetCmd = &cobra.Command{
	Use:   "get <dashboard_url_path> <view_index> <badge_index>",
	Short: "Get a specific badge",
	Long:  `Get a specific badge from a view by index.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runBadgeGet,
}

func init() {
	dashboardBadgeCmd.AddCommand(badgeGetCmd)
}

func runBadgeGet(cmd *cobra.Command, args []string) error {
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

	badge := badges[badgeIndex]
	badgeData := make(map[string]interface{})
	switch b := badge.(type) {
	case map[string]interface{}:
		for k, val := range b {
			badgeData[k] = val
		}
	case string:
		badgeData["entity"] = b
	}
	badgeData["index"] = badgeIndex

	client.PrintOutput(badgeData, textMode, "")
	return nil
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var badgeDeleteForce bool

var badgeDeleteCmd = &cobra.Command{
	Use:   "delete <dashboard_url_path> <view_index> <badge_index>",
	Short: "Delete a badge",
	Long:  `Delete a badge from a view by index.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runBadgeDelete,
}

func init() {
	dashboardBadgeCmd.AddCommand(badgeDeleteCmd)
	badgeDeleteCmd.Flags().BoolVarP(&badgeDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runBadgeDelete(cmd *cobra.Command, args []string) error {
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

	// Confirmation prompt
	if !badgeDeleteForce && !textMode {
		fmt.Printf("Are you sure you want to delete badge at index %d? [y/N]: ", badgeIndex)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
	}

	// Remove the badge
	badges = append(badges[:badgeIndex], badges[badgeIndex+1:]...)
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

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Badge at index %d deleted.", badgeIndex))
	return nil
}

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

var viewDeleteForce bool

var viewDeleteCmd = &cobra.Command{
	Use:   "delete <dashboard_url_path> <view_index>",
	Short: "Delete a view",
	Long:  `Delete a view from a dashboard by index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runViewDelete,
}

func init() {
	dashboardViewCmd.AddCommand(viewDeleteCmd)
	viewDeleteCmd.Flags().BoolVarP(&viewDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runViewDelete(cmd *cobra.Command, args []string) error {
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

	// Get view title for confirmation
	viewTitle := fmt.Sprintf("view at index %d", viewIndex)
	if viewMap, ok := views[viewIndex].(map[string]interface{}); ok {
		if title, ok := viewMap["title"].(string); ok {
			viewTitle = fmt.Sprintf("view '%s' (index %d)", title, viewIndex)
		}
	}

	// Confirmation prompt
	if !viewDeleteForce && !textMode {
		fmt.Printf("Are you sure you want to delete %s? [y/N]: ", viewTitle)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
	}

	// Remove the view
	views = append(views[:viewIndex], views[viewIndex+1:]...)
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

	client.PrintSuccess(nil, textMode, fmt.Sprintf("View at index %d deleted.", viewIndex))
	return nil
}

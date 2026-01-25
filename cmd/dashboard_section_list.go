package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sectionListCmd = &cobra.Command{
	Use:   "list <dashboard_url_path> <view_index>",
	Short: "List sections in a view",
	Long:  `List all sections in a dashboard view.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runSectionList,
}

func init() {
	dashboardSectionCmd.AddCommand(sectionListCmd)
}

func runSectionList(cmd *cobra.Command, args []string) error {
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
		client.PrintOutput([]interface{}{}, textMode, "")
		return nil
	}

	// Add index to each section for easier reference
	sectionList := make([]map[string]interface{}, len(sections))
	for i, s := range sections {
		sectionData := make(map[string]interface{})
		if sectionMap, ok := s.(map[string]interface{}); ok {
			for k, val := range sectionMap {
				sectionData[k] = val
			}
		}
		sectionData["index"] = i
		sectionList[i] = sectionData
	}

	client.PrintOutput(sectionList, textMode, "")
	return nil
}

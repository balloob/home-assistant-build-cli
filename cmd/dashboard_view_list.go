package cmd

import (
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var viewListCmd = &cobra.Command{
	Use:   "list <dashboard_url_path>",
	Short: "List views in a dashboard",
	Long:  `List all views in a dashboard.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runViewList,
}

func init() {
	dashboardViewCmd.AddCommand(viewListCmd)
}

func runViewList(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
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
		client.PrintOutput([]interface{}{}, textMode, "")
		return nil
	}

	views, ok := config["views"].([]interface{})
	if !ok {
		client.PrintOutput([]interface{}{}, textMode, "")
		return nil
	}

	// Add index to each view for easier reference
	viewList := make([]map[string]interface{}, len(views))
	for i, v := range views {
		if viewMap, ok := v.(map[string]interface{}); ok {
			viewData := make(map[string]interface{})
			for k, val := range viewMap {
				viewData[k] = val
			}
			viewData["index"] = i
			viewList[i] = viewData
		}
	}

	client.PrintOutput(viewList, textMode, "")
	return nil
}

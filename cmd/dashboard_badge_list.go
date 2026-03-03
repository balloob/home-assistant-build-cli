package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var badgeListCmd = &cobra.Command{
	Use:   "list <dashboard_url_path> <view_index>",
	Short: "List badges in a view",
	Long:  `List all badges in a dashboard view.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runBadgeList,
}

func init() {
	dashboardBadgeCmd.AddCommand(badgeListCmd)
}

func runBadgeList(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	viewIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid view index: %s", args[1])
	}

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
		output.PrintOutput([]interface{}{}, textMode, "")
		return nil
	}

	// Add index to each badge for easier reference
	badgeList := make([]map[string]interface{}, len(badges))
	for i, b := range badges {
		badgeData := make(map[string]interface{})
		switch badge := b.(type) {
		case map[string]interface{}:
			for k, val := range badge {
				badgeData[k] = val
			}
		case string:
			// Simple entity_id badge
			badgeData["entity"] = badge
		}
		badgeData["index"] = i
		badgeList[i] = badgeData
	}

	output.PrintOutput(badgeList, textMode, "")
	return nil
}

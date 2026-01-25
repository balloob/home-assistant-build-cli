package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sectionGetCmd = &cobra.Command{
	Use:   "get <dashboard_url_path> <view_index> <section_index>",
	Short: "Get a specific section",
	Long:  `Get a specific section from a view by index.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runSectionGet,
}

func init() {
	dashboardSectionCmd.AddCommand(sectionGetCmd)
}

func runSectionGet(cmd *cobra.Command, args []string) error {
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

	section := sections[sectionIndex]
	if sectionMap, ok := section.(map[string]interface{}); ok {
		sectionMap["index"] = sectionIndex
	}

	client.PrintOutput(section, textMode, "")
	return nil
}

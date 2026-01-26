package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	sectionGetDashboard    string
	sectionGetViewIndex    int
	sectionGetSectionIndex int
)

var sectionGetCmd = &cobra.Command{
	Use:   "get [dashboard_url_path] [view_index] [section_index]",
	Short: "Get a specific section",
	Long:  `Get a specific section from a view by index.`,
	Args:  cobra.MaximumNArgs(3),
	RunE:  runSectionGet,
}

func init() {
	dashboardSectionCmd.AddCommand(sectionGetCmd)
	sectionGetCmd.Flags().StringVar(&sectionGetDashboard, "dashboard", "", "Dashboard URL path")
	sectionGetCmd.Flags().IntVar(&sectionGetViewIndex, "view", -1, "View index")
	sectionGetCmd.Flags().IntVar(&sectionGetSectionIndex, "index", -1, "Section index")
}

func runSectionGet(cmd *cobra.Command, args []string) error {
	urlPath := sectionGetDashboard
	if urlPath == "" && len(args) > 0 {
		urlPath = args[0]
	}
	if urlPath == "" {
		return fmt.Errorf("dashboard URL path is required (use --dashboard flag or first positional argument)")
	}

	viewIndex := sectionGetViewIndex
	if viewIndex < 0 && len(args) > 1 {
		var err error
		viewIndex, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid view index: %s", args[1])
		}
	}
	if viewIndex < 0 {
		return fmt.Errorf("view index is required (use --view flag or second positional argument)")
	}

	sectionIndex := sectionGetSectionIndex
	if sectionIndex < 0 && len(args) > 2 {
		var err error
		sectionIndex, err = strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid section index: %s", args[2])
		}
	}
	if sectionIndex < 0 {
		return fmt.Errorf("section index is required (use --index flag or third positional argument)")
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

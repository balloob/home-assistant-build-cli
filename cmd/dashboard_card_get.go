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
	cardGetDashboard string
	cardGetView      int
	cardGetIndex     int
	cardGetSection   int
)

var cardGetCmd = &cobra.Command{
	Use:   "get [dashboard_url_path] [view_index] [card_index]",
	Short: "Get a specific card",
	Long: `Get a specific card from a section by index.

If section is not specified, uses the last section.`,
	Args: cobra.MaximumNArgs(3),
	RunE: runCardGet,
}

func init() {
	dashboardCardCmd.AddCommand(cardGetCmd)
	cardGetCmd.Flags().StringVar(&cardGetDashboard, "dashboard", "", "Dashboard URL path")
	cardGetCmd.Flags().IntVar(&cardGetView, "view", -1, "View index")
	cardGetCmd.Flags().IntVar(&cardGetIndex, "index", -1, "Card index")
	cardGetCmd.Flags().IntVarP(&cardGetSection, "section", "s", -1, "Section index (if card is in a section)")
}

func runCardGet(cmd *cobra.Command, args []string) error {
	urlPath := cardGetDashboard
	if urlPath == "" && len(args) > 0 {
		urlPath = args[0]
	}
	if urlPath == "" {
		return fmt.Errorf("dashboard URL path is required (use --dashboard flag or first positional argument)")
	}

	viewIndex := cardGetView
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

	cardIndex := cardGetIndex
	if cardIndex < 0 && len(args) > 2 {
		var err error
		cardIndex, err = strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid card index: %s", args[2])
		}
	}
	if cardIndex < 0 {
		return fmt.Errorf("card index is required (use --index flag or third positional argument)")
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

	// Get sections
	sections, _ := view["sections"].([]interface{})
	if sections == nil || len(sections) == 0 {
		return fmt.Errorf("no sections in view")
	}

	// Determine section index: use provided value or default to last section
	sectionIndex := cardGetSection
	if sectionIndex < 0 {
		sectionIndex = len(sections) - 1
	}

	if sectionIndex >= len(sections) {
		return fmt.Errorf("section index %d out of range (0-%d)", sectionIndex, len(sections)-1)
	}

	section, ok := sections[sectionIndex].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid section at index %d", sectionIndex)
	}

	cards, _ := section["cards"].([]interface{})
	if cards == nil {
		return fmt.Errorf("no cards found")
	}

	if cardIndex < 0 || cardIndex >= len(cards) {
		return fmt.Errorf("card index %d out of range (0-%d)", cardIndex, len(cards)-1)
	}

	card := cards[cardIndex]
	if cardMap, ok := card.(map[string]interface{}); ok {
		cardMap["index"] = cardIndex
	}

	client.PrintOutput(card, textMode, "")
	return nil
}

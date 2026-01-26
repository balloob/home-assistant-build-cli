package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cardGetSection int

var cardGetCmd = &cobra.Command{
	Use:   "get <dashboard_url_path> <view_index> <card_index>",
	Short: "Get a specific card",
	Long: `Get a specific card from a section by index.

If section is not specified, uses the last section.`,
	Args: cobra.ExactArgs(3),
	RunE: runCardGet,
}

func init() {
	dashboardCardCmd.AddCommand(cardGetCmd)
	cardGetCmd.Flags().IntVarP(&cardGetSection, "section", "s", -1, "Section index (if card is in a section)")
}

func runCardGet(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	viewIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid view index: %s", args[1])
	}
	cardIndex, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid card index: %s", args[2])
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

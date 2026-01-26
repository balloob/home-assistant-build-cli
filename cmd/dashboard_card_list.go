package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cardListSection int

var cardListCmd = &cobra.Command{
	Use:   "list <dashboard_url_path> <view_index>",
	Short: "List cards in a section",
	Long: `List all cards in a dashboard section.

If section is not specified, uses the last section.`,
	Args: cobra.ExactArgs(2),
	RunE: runCardList,
}

func init() {
	dashboardCardCmd.AddCommand(cardListCmd)
	cardListCmd.Flags().IntVarP(&cardListSection, "section", "s", -1, "Section index (if cards are in a section)")
}

func runCardList(cmd *cobra.Command, args []string) error {
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

	// Get sections
	sections, _ := view["sections"].([]interface{})
	if sections == nil || len(sections) == 0 {
		return fmt.Errorf("no sections in view")
	}

	// Determine section index: use provided value or default to last section
	sectionIndex := cardListSection
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
		client.PrintOutput([]interface{}{}, textMode, "")
		return nil
	}

	// Add index to each card for easier reference
	cardList := make([]map[string]interface{}, len(cards))
	for i, c := range cards {
		cardData := make(map[string]interface{})
		if cardMap, ok := c.(map[string]interface{}); ok {
			for k, val := range cardMap {
				cardData[k] = val
			}
		}
		cardData["index"] = i
		cardList[i] = cardData
	}

	client.PrintOutput(cardList, textMode, "")
	return nil
}

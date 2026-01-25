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
	Short: "List cards in a view or section",
	Long:  `List all cards in a dashboard view or section. Use --section to specify a section index.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runCardList,
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

	var cards []interface{}

	if cardListSection >= 0 {
		// Get cards from section
		sections, ok := view["sections"].([]interface{})
		if !ok {
			return fmt.Errorf("no sections in view")
		}
		if cardListSection >= len(sections) {
			return fmt.Errorf("section index %d out of range (0-%d)", cardListSection, len(sections)-1)
		}
		section, ok := sections[cardListSection].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid section at index %d", cardListSection)
		}
		cards, _ = section["cards"].([]interface{})
	} else {
		// Get cards directly from view
		cards, _ = view["cards"].([]interface{})
	}

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

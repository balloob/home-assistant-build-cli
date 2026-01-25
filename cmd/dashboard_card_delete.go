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

var (
	cardDeleteForce   bool
	cardDeleteSection int
)

var cardDeleteCmd = &cobra.Command{
	Use:   "delete <dashboard_url_path> <view_index> <card_index>",
	Short: "Delete a card",
	Long:  `Delete a card from a view or section by index.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runCardDelete,
}

func init() {
	dashboardCardCmd.AddCommand(cardDeleteCmd)
	cardDeleteCmd.Flags().BoolVarP(&cardDeleteForce, "force", "f", false, "Skip confirmation prompt")
	cardDeleteCmd.Flags().IntVarP(&cardDeleteSection, "section", "s", -1, "Section index (if card is in a section)")
}

func runCardDelete(cmd *cobra.Command, args []string) error {
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

	view, ok := views[viewIndex].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid view at index %d", viewIndex)
	}

	var cards []interface{}
	var section map[string]interface{}
	var sections []interface{}

	if cardDeleteSection >= 0 {
		// Get cards from section
		sections, ok = view["sections"].([]interface{})
		if !ok {
			return fmt.Errorf("no sections in view")
		}
		if cardDeleteSection >= len(sections) {
			return fmt.Errorf("section index %d out of range (0-%d)", cardDeleteSection, len(sections)-1)
		}
		section, ok = sections[cardDeleteSection].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid section at index %d", cardDeleteSection)
		}
		cards, _ = section["cards"].([]interface{})
	} else {
		// Get cards directly from view
		cards, _ = view["cards"].([]interface{})
	}

	if cards == nil {
		return fmt.Errorf("no cards found")
	}

	if cardIndex < 0 || cardIndex >= len(cards) {
		return fmt.Errorf("card index %d out of range (0-%d)", cardIndex, len(cards)-1)
	}

	// Get card type for confirmation
	cardDesc := fmt.Sprintf("card at index %d", cardIndex)
	if cardMap, ok := cards[cardIndex].(map[string]interface{}); ok {
		if cardType, ok := cardMap["type"].(string); ok {
			cardDesc = fmt.Sprintf("%s card (index %d)", cardType, cardIndex)
		}
	}

	// Confirmation prompt
	if !cardDeleteForce && !textMode {
		fmt.Printf("Are you sure you want to delete %s? [y/N]: ", cardDesc)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
	}

	// Remove the card
	cards = append(cards[:cardIndex], cards[cardIndex+1:]...)

	if cardDeleteSection >= 0 {
		section["cards"] = cards
		sections[cardDeleteSection] = section
		view["sections"] = sections
	} else {
		view["cards"] = cards
	}

	views[viewIndex] = view
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

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Card at index %d deleted.", cardIndex))
	return nil
}

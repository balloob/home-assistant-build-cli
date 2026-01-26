package cmd

import (
	"fmt"
	"strconv"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cardUpdateData    string
	cardUpdateFile    string
	cardUpdateFormat  string
	cardUpdateType    string
	cardUpdateEntity  string
	cardUpdateSection int
)

var cardUpdateCmd = &cobra.Command{
	Use:   "update <dashboard_url_path> <view_index> <card_index>",
	Short: "Update a card",
	Long: `Update a card in a section by index.

If section is not specified, uses the last section.`,
	Args: cobra.ExactArgs(3),
	RunE: runCardUpdate,
}

func init() {
	dashboardCardCmd.AddCommand(cardUpdateCmd)
	cardUpdateCmd.Flags().StringVarP(&cardUpdateData, "data", "d", "", "Card configuration as JSON (replaces entire card)")
	cardUpdateCmd.Flags().StringVarP(&cardUpdateFile, "file", "f", "", "Path to config file")
	cardUpdateCmd.Flags().StringVar(&cardUpdateFormat, "format", "", "Input format (json, yaml)")
	cardUpdateCmd.Flags().StringVar(&cardUpdateType, "type", "", "Card type")
	cardUpdateCmd.Flags().StringVar(&cardUpdateEntity, "entity", "", "Entity ID")
	cardUpdateCmd.Flags().IntVarP(&cardUpdateSection, "section", "s", -1, "Section index (if card is in a section)")
}

func runCardUpdate(cmd *cobra.Command, args []string) error {
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

	// Get sections
	sections, _ := view["sections"].([]interface{})
	if sections == nil || len(sections) == 0 {
		return fmt.Errorf("no sections in view")
	}

	// Determine section index: use provided value or default to last section
	sectionIndex := cardUpdateSection
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

	// Get existing card
	existingCard, ok := cards[cardIndex].(map[string]interface{})
	if !ok {
		existingCard = make(map[string]interface{})
	}

	// If data or file provided, replace the entire card config
	if cardUpdateData != "" || cardUpdateFile != "" {
		newConfig, err := input.ParseInput(cardUpdateData, cardUpdateFile, cardUpdateFormat)
		if err != nil {
			return err
		}
		existingCard = newConfig
	}

	// Apply flag updates
	if cmd.Flags().Changed("type") {
		existingCard["type"] = cardUpdateType
	}
	if cmd.Flags().Changed("entity") {
		existingCard["entity"] = cardUpdateEntity
	}

	cards[cardIndex] = existingCard
	section["cards"] = cards
	sections[sectionIndex] = section
	view["sections"] = sections

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

	existingCard["index"] = cardIndex
	client.PrintSuccess(existingCard, textMode, fmt.Sprintf("Card at index %d updated.", cardIndex))
	return nil
}

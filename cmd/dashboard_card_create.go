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
	cardCreateData    string
	cardCreateFile    string
	cardCreateFormat  string
	cardCreateType    string
	cardCreateEntity  string
	cardCreateSection int
)

var cardCreateCmd = &cobra.Command{
	Use:   "create <dashboard_url_path> <view_index>",
	Short: "Create a new card",
	Long:  `Create a new card in a dashboard view or section.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runCardCreate,
}

func init() {
	dashboardCardCmd.AddCommand(cardCreateCmd)
	cardCreateCmd.Flags().StringVarP(&cardCreateData, "data", "d", "", "Card configuration as JSON")
	cardCreateCmd.Flags().StringVarP(&cardCreateFile, "file", "f", "", "Path to config file")
	cardCreateCmd.Flags().StringVar(&cardCreateFormat, "format", "", "Input format (json, yaml)")
	cardCreateCmd.Flags().StringVar(&cardCreateType, "type", "", "Card type (e.g., entities, button, markdown)")
	cardCreateCmd.Flags().StringVar(&cardCreateEntity, "entity", "", "Entity ID (for simple entity cards)")
	cardCreateCmd.Flags().IntVarP(&cardCreateSection, "section", "s", -1, "Section index (if card should be in a section)")
}

func runCardCreate(cmd *cobra.Command, args []string) error {
	urlPath := args[0]
	viewIndex, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid view index: %s", args[1])
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	var cardConfig map[string]interface{}

	// If data or file provided, parse it
	if cardCreateData != "" || cardCreateFile != "" {
		cardConfig, err = input.ParseInput(cardCreateData, cardCreateFile, cardCreateFormat)
		if err != nil {
			return err
		}
	} else {
		cardConfig = make(map[string]interface{})
	}

	// Apply flags if provided
	if cardCreateType != "" {
		cardConfig["type"] = cardCreateType
	}
	if cardCreateEntity != "" {
		cardConfig["entity"] = cardCreateEntity
	}

	// Ensure type is set
	if _, ok := cardConfig["type"]; !ok {
		return fmt.Errorf("card type is required (use --type or provide in data)")
	}

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
	var newCardIndex int

	if cardCreateSection >= 0 {
		// Add card to section
		sections, ok := view["sections"].([]interface{})
		if !ok {
			return fmt.Errorf("no sections in view")
		}
		if cardCreateSection >= len(sections) {
			return fmt.Errorf("section index %d out of range (0-%d)", cardCreateSection, len(sections)-1)
		}
		section, ok := sections[cardCreateSection].(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid section at index %d", cardCreateSection)
		}

		cards, _ = section["cards"].([]interface{})
		if cards == nil {
			cards = []interface{}{}
		}
		cards = append(cards, cardConfig)
		newCardIndex = len(cards) - 1
		section["cards"] = cards
		sections[cardCreateSection] = section
		view["sections"] = sections
	} else {
		// Add card directly to view
		cards, _ = view["cards"].([]interface{})
		if cards == nil {
			cards = []interface{}{}
		}
		cards = append(cards, cardConfig)
		newCardIndex = len(cards) - 1
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

	cardConfig["index"] = newCardIndex
	location := "view"
	if cardCreateSection >= 0 {
		location = fmt.Sprintf("section %d", cardCreateSection)
	}
	client.PrintSuccess(cardConfig, textMode, fmt.Sprintf("Card created at index %d in %s.", newCardIndex, location))
	return nil
}

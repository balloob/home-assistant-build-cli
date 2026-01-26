package cmd

import (
	"fmt"
	"strconv"
	"strings"

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
	Use:   "create <dashboard_url_path> [view_index]",
	Short: "Create a new card",
	Long: `Create a new card in a dashboard view or section.

If view_index is not specified, uses the last view. If no views exist, creates one.
If section is not specified, uses the last section. If no sections exist, creates one.
If type is not specified, defaults to "tile".`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runCardCreate,
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

	// View index is optional - will default to last view
	viewIndex := -1
	if len(args) > 1 {
		var err error
		viewIndex, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid view index: %s", args[1])
		}
	}

	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	var cardConfig map[string]interface{}
	var err error

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

	// Default type to "tile" if not set
	if _, ok := cardConfig["type"]; !ok {
		cardConfig["type"] = "tile"
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
		config = map[string]interface{}{
			"views": []interface{}{},
		}
	}

	views, ok := config["views"].([]interface{})
	if !ok {
		views = []interface{}{}
	}

	// If no views exist, create one
	viewCreated := false
	if len(views) == 0 {
		newView := map[string]interface{}{
			"title":    "Home",
			"sections": []interface{}{},
		}
		views = append(views, newView)
		viewCreated = true
	}

	// Default to last view if not specified
	if viewIndex < 0 {
		viewIndex = len(views) - 1
	}

	if viewIndex >= len(views) {
		return fmt.Errorf("view index %d out of range (0-%d)", viewIndex, len(views)-1)
	}

	view, ok := views[viewIndex].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid view at index %d", viewIndex)
	}

	// Get or create sections
	sections, _ := view["sections"].([]interface{})
	if sections == nil {
		sections = []interface{}{}
	}

	// If no sections exist, create one
	sectionCreated := false
	if len(sections) == 0 {
		newSection := map[string]interface{}{
			"type":  "grid",
			"cards": []interface{}{},
		}
		sections = append(sections, newSection)
		view["sections"] = sections
		sectionCreated = true
	}

	// Determine section index: use provided value or default to last section
	sectionIndex := cardCreateSection
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
		cards = []interface{}{}
	}
	cards = append(cards, cardConfig)
	newCardIndex := len(cards) - 1
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

	cardConfig["index"] = newCardIndex

	// Build descriptive message
	var msgParts []string
	if viewCreated {
		msgParts = append(msgParts, fmt.Sprintf("view %d created", viewIndex))
	}
	if sectionCreated {
		msgParts = append(msgParts, fmt.Sprintf("section %d created", sectionIndex))
	}
	msgParts = append(msgParts, fmt.Sprintf("card created at index %d in view %d section %d", newCardIndex, viewIndex, sectionIndex))

	var msg string
	if len(msgParts) > 1 {
		msg = ""
		for i, part := range msgParts {
			if i == 0 {
				msg = strings.ToUpper(part[:1]) + part[1:]
			} else {
				msg += ", " + part
			}
		}
		msg += "."
	} else {
		msg = "Card created at index " + strconv.Itoa(newCardIndex) + " in view " + strconv.Itoa(viewIndex) + " section " + strconv.Itoa(sectionIndex) + "."
	}

	client.PrintSuccess(cardConfig, textMode, msg)
	return nil
}

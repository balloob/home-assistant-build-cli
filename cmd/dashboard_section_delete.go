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

var sectionDeleteForce bool

var sectionDeleteCmd = &cobra.Command{
	Use:   "delete <dashboard_url_path> <view_index> <section_index>",
	Short: "Delete a section",
	Long:  `Delete a section from a view by index.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runSectionDelete,
}

func init() {
	dashboardSectionCmd.AddCommand(sectionDeleteCmd)
	sectionDeleteCmd.Flags().BoolVarP(&sectionDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runSectionDelete(cmd *cobra.Command, args []string) error {
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

	sections, ok := view["sections"].([]interface{})
	if !ok {
		return fmt.Errorf("no sections in view")
	}

	if sectionIndex < 0 || sectionIndex >= len(sections) {
		return fmt.Errorf("section index %d out of range (0-%d)", sectionIndex, len(sections)-1)
	}

	// Get section title for confirmation
	sectionTitle := fmt.Sprintf("section at index %d", sectionIndex)
	if sectionMap, ok := sections[sectionIndex].(map[string]interface{}); ok {
		if title, ok := sectionMap["title"].(string); ok && title != "" {
			sectionTitle = fmt.Sprintf("section '%s' (index %d)", title, sectionIndex)
		}
	}

	// Confirmation prompt
	if !sectionDeleteForce && !textMode {
		fmt.Printf("Are you sure you want to delete %s? [y/N]: ", sectionTitle)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("deletion cancelled")
		}
	}

	// Remove the section
	sections = append(sections[:sectionIndex], sections[sectionIndex+1:]...)
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

	client.PrintSuccess(nil, textMode, fmt.Sprintf("Section at index %d deleted.", sectionIndex))
	return nil
}

package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperLocalCalendarListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all local calendar helpers",
	Long:  `List all local calendar helpers created via config entries.`,
	RunE:  runHelperLocalCalendarList,
}

func init() {
	helperLocalCalendarParentCmd.AddCommand(helperLocalCalendarListCmd)
}

func runHelperLocalCalendarList(cmd *cobra.Command, args []string) error {
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

	// Get config entries for the "local_calendar" domain
	entries, err := ws.ConfigEntriesList("local_calendar")
	if err != nil {
		return err
	}

	var result []map[string]interface{}
	for _, e := range entries {
		entry, ok := e.(map[string]interface{})
		if !ok {
			continue
		}

		item := map[string]interface{}{
			"entry_id": entry["entry_id"],
			"title":    entry["title"],
		}

		// Extract domain from entry if available
		if domain, ok := entry["domain"].(string); ok {
			item["domain"] = domain
		}

		result = append(result, item)
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

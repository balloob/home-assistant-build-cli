package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperTemplateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all template entities",
	Long:  `List all template entity helpers created via config entries.`,
	RunE:  runHelperTemplateList,
}

var (
	helperTemplateListCount bool
	helperTemplateListBrief bool
	helperTemplateListLimit int
)

func init() {
	helperTemplateParentCmd.AddCommand(helperTemplateListCmd)
	helperTemplateListCmd.Flags().BoolVarP(&helperTemplateListCount, "count", "c", false, "Return only the count of items")
	helperTemplateListCmd.Flags().BoolVarP(&helperTemplateListBrief, "brief", "b", false, "Return minimal fields (entry_id and title only)")
	helperTemplateListCmd.Flags().IntVarP(&helperTemplateListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperTemplateList(cmd *cobra.Command, args []string) error {
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

	// Get config entries for the "template" domain
	entries, err := ws.ConfigEntriesList("template")
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

	// Handle count mode
	if helperTemplateListCount {
		client.PrintOutput(map[string]interface{}{"count": len(result)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperTemplateListLimit > 0 && len(result) > helperTemplateListLimit {
		result = result[:helperTemplateListLimit]
	}

	// Handle brief mode
	if helperTemplateListBrief {
		var brief []map[string]interface{}
		for _, item := range result {
			brief = append(brief, map[string]interface{}{
				"entry_id": item["entry_id"],
				"title":    item["title"],
			})
		}
		client.PrintOutput(brief, textMode, "")
		return nil
	}

	client.PrintOutput(result, textMode, "")
	return nil
}

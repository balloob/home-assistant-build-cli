package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperLocalTodoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all local to-do list helpers",
	Long:  `List all local to-do list helpers created via config entries.`,
	RunE:  runHelperLocalTodoList,
}

var (
	helperLocalTodoListCount bool
	helperLocalTodoListBrief bool
	helperLocalTodoListLimit int
)

func init() {
	helperLocalTodoParentCmd.AddCommand(helperLocalTodoListCmd)
	helperLocalTodoListCmd.Flags().BoolVarP(&helperLocalTodoListCount, "count", "c", false, "Return only the count of items")
	helperLocalTodoListCmd.Flags().BoolVarP(&helperLocalTodoListBrief, "brief", "b", false, "Return minimal fields (entry_id and title only)")
	helperLocalTodoListCmd.Flags().IntVarP(&helperLocalTodoListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperLocalTodoList(cmd *cobra.Command, args []string) error {
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

	// Get config entries for the "local_todo" domain
	entries, err := ws.ConfigEntriesList("local_todo")
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
	if helperLocalTodoListCount {
		client.PrintOutput(map[string]interface{}{"count": len(result)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperLocalTodoListLimit > 0 && len(result) > helperLocalTodoListLimit {
		result = result[:helperLocalTodoListLimit]
	}

	// Handle brief mode
	if helperLocalTodoListBrief {
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

package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var helperUtilityMeterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all utility meters",
	Long:  `List all utility meter helpers created via config entries.`,
	RunE:  runHelperUtilityMeterList,
}

var (
	helperUtilityMeterListCount bool
	helperUtilityMeterListBrief bool
	helperUtilityMeterListLimit int
)

func init() {
	helperUtilityMeterParentCmd.AddCommand(helperUtilityMeterListCmd)
	helperUtilityMeterListCmd.Flags().BoolVarP(&helperUtilityMeterListCount, "count", "c", false, "Return only the count of items")
	helperUtilityMeterListCmd.Flags().BoolVarP(&helperUtilityMeterListBrief, "brief", "b", false, "Return minimal fields (entry_id and title only)")
	helperUtilityMeterListCmd.Flags().IntVarP(&helperUtilityMeterListLimit, "limit", "n", 0, "Limit results to N items")
}

func runHelperUtilityMeterList(cmd *cobra.Command, args []string) error {
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

	entries, err := ws.ConfigEntriesList("utility_meter")
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

		if domain, ok := entry["domain"].(string); ok {
			item["domain"] = domain
		}

		result = append(result, item)
	}

	// Handle count mode
	if helperUtilityMeterListCount {
		client.PrintOutput(map[string]interface{}{"count": len(result)}, textMode, "")
		return nil
	}

	// Apply limit
	if helperUtilityMeterListLimit > 0 && len(result) > helperUtilityMeterListLimit {
		result = result[:helperUtilityMeterListLimit]
	}

	// Handle brief mode
	if helperUtilityMeterListBrief {
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

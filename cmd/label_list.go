package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var labelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all labels",
	Long:  `List all labels in Home Assistant.`,
	RunE:  runLabelList,
}

var (
	labelListCount bool
	labelListBrief bool
	labelListLimit int
)

func init() {
	labelCmd.AddCommand(labelListCmd)
	labelListCmd.Flags().BoolVarP(&labelListCount, "count", "c", false, "Return only the count of items")
	labelListCmd.Flags().BoolVarP(&labelListBrief, "brief", "b", false, "Return minimal fields (label_id and name only)")
	labelListCmd.Flags().IntVarP(&labelListLimit, "limit", "n", 0, "Limit results to N items")
}

func runLabelList(cmd *cobra.Command, args []string) error {
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

	labels, err := ws.LabelRegistryList()
	if err != nil {
		return err
	}

	// Handle count mode
	if labelListCount {
		client.PrintOutput(map[string]interface{}{"count": len(labels)}, textMode, "")
		return nil
	}

	// Apply limit
	if labelListLimit > 0 && len(labels) > labelListLimit {
		labels = labels[:labelListLimit]
	}

	// Handle brief mode
	if labelListBrief {
		var brief []map[string]interface{}
		for _, l := range labels {
			if label, ok := l.(map[string]interface{}); ok {
				brief = append(brief, map[string]interface{}{
					"label_id": label["label_id"],
					"name":     label["name"],
				})
			}
		}
		client.PrintOutput(brief, textMode, "")
		return nil
	}

	client.PrintOutput(labels, textMode, "")
	return nil
}

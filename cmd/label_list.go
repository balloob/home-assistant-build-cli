package cmd

import (
	"fmt"

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
	labelListID    string
	labelListCount bool
	labelListBrief bool
	labelListLimit int
)

func init() {
	labelCmd.AddCommand(labelListCmd)
	labelListCmd.Flags().StringVar(&labelListID, "label-id", "", "Filter by label ID")
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

	// Apply label ID filter
	if labelListID != "" {
		var filtered []interface{}
		for _, l := range labels {
			if label, ok := l.(map[string]interface{}); ok {
				labelID, _ := label["label_id"].(string)
				if labelID == labelListID {
					filtered = append(filtered, l)
				}
			}
		}
		labels = filtered
	}

	// Handle count mode
	if labelListCount {
		if textMode {
			fmt.Printf("Count: %d\n", len(labels))
		} else {
			client.PrintOutput(map[string]interface{}{"count": len(labels)}, false, "")
		}
		return nil
	}

	// Apply limit
	if labelListLimit > 0 && len(labels) > labelListLimit {
		labels = labels[:labelListLimit]
	}

	// Handle brief mode
	if labelListBrief {
		if textMode {
			for _, l := range labels {
				if label, ok := l.(map[string]interface{}); ok {
					name, _ := label["name"].(string)
					labelID, _ := label["label_id"].(string)
					fmt.Printf("%s (%s)\n", name, labelID)
				}
			}
		} else {
			var brief []map[string]interface{}
			for _, l := range labels {
				if label, ok := l.(map[string]interface{}); ok {
					brief = append(brief, map[string]interface{}{
						"label_id": label["label_id"],
						"name":     label["name"],
					})
				}
			}
			client.PrintOutput(brief, false, "")
		}
		return nil
	}

	// Full output
	if textMode {
		if len(labels) == 0 {
			fmt.Println("No labels.")
			return nil
		}
		for _, l := range labels {
			if label, ok := l.(map[string]interface{}); ok {
				name, _ := label["name"].(string)
				labelID, _ := label["label_id"].(string)
				color, _ := label["color"].(string)

				if color != "" {
					fmt.Printf("%s (%s): %s\n", name, labelID, color)
				} else {
					fmt.Printf("%s (%s)\n", name, labelID)
				}
			}
		}
	} else {
		client.PrintOutput(labels, false, "")
	}
	return nil
}

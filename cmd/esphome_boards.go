package cmd

import (
	"strings"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var esphomeBoardsCmd = &cobra.Command{
	Use:   "boards <platform>",
	Short: "List supported boards for an ESPHome platform",
	Long:  `List the available board IDs for a given ESPHome platform such as esp32, esp8266, or rp2040.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()
		esClient, err := getESPHomeClient()
		if err != nil {
			return err
		}

		platform := strings.ToLower(args[0])
		boards, err := esClient.GetBoards(platform)
		if err != nil {
			return err
		}

		result := make([]map[string]any, 0, len(boards))
		for _, board := range boards {
			result = append(result, map[string]any{
				"id":       board.ID,
				"name":     board.Name,
				"platform": platform,
			})
		}

		output.PrintOutput(result, textMode, "")
		return nil
	},
}

func init() {
	esphomeCmd.AddCommand(esphomeBoardsCmd)
}

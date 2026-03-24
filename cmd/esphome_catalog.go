package cmd

import (
	"context"

	"github.com/home-assistant/hab/internal/esphomecatalog"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	esphomeCatalogRef              string
	esphomeCatalogSearchBoard      string
	esphomeCatalogSearchType       string
	esphomeCatalogSearchDifficulty int
	esphomeCatalogSearchLimit      int
	esphomeCatalogShowIncludeYAML  bool
)

var esphomeCatalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Browse ESPHome community device catalog",
	Long:  `Search and inspect device templates from the esphome-devices catalog.`,
}

var esphomeCatalogSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search catalog devices by slug and metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()
		catalogClient := esphomecatalog.NewClient(esphomeCatalogRef)

		results, err := catalogClient.Search(context.Background(), esphomecatalog.SearchOptions{
			Query:      args[0],
			Board:      esphomeCatalogSearchBoard,
			DeviceType: esphomeCatalogSearchType,
			Difficulty: esphomeCatalogSearchDifficulty,
			Limit:      esphomeCatalogSearchLimit,
		})
		if err != nil {
			return err
		}

		output.PrintOutput(results, textMode, "")
		return nil
	},
}

var esphomeCatalogShowCmd = &cobra.Command{
	Use:   "show <slug>",
	Short: "Show catalog metadata and source config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()
		catalogClient := esphomecatalog.NewClient(esphomeCatalogRef)

		device, err := catalogClient.GetDevice(context.Background(), args[0])
		if err != nil {
			return err
		}

		data := map[string]any{
			"slug":               device.Slug,
			"title":              device.Title,
			"type":               device.DeviceType,
			"board":              device.Board,
			"difficulty":         device.Difficulty,
			"project_url":        device.ProjectURL,
			"project_name":       device.ProjectName,
			"package_import_url": device.PackageImportURL,
			"yaml_path":          device.YAMLPath,
			"content_source":     device.ContentSource,
			"markdown_path":      device.MarkdownPath,
		}
		if esphomeCatalogShowIncludeYAML {
			data["yaml_content"] = device.YAMLContent
		}

		output.PrintOutput(data, textMode, "")
		return nil
	},
}

func init() {
	esphomeCmd.AddCommand(esphomeCatalogCmd)

	esphomeCatalogCmd.PersistentFlags().StringVar(&esphomeCatalogRef, "ref", "main", "Catalog git ref or branch")

	esphomeCatalogCmd.AddCommand(esphomeCatalogSearchCmd)
	esphomeCatalogSearchCmd.Flags().StringVar(&esphomeCatalogSearchBoard, "board", "", "Filter by board family (esp32, esp8266, rp2040, ...)")
	esphomeCatalogSearchCmd.Flags().StringVar(&esphomeCatalogSearchType, "type", "", "Filter by device type (plug, relay, light, ...)")
	esphomeCatalogSearchCmd.Flags().IntVar(&esphomeCatalogSearchDifficulty, "difficulty", 0, "Filter by difficulty 1-5")
	esphomeCatalogSearchCmd.Flags().IntVar(&esphomeCatalogSearchLimit, "limit", 20, "Max results to return")

	esphomeCatalogCmd.AddCommand(esphomeCatalogShowCmd)
	esphomeCatalogShowCmd.Flags().BoolVar(&esphomeCatalogShowIncludeYAML, "include-yaml", false, "Include the source YAML content in output")
}

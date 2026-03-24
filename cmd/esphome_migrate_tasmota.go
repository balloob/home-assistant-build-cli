package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/esphomeconfig"
	"github.com/home-assistant/hab/internal/esphometasmota"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	esphomeMigrateTemplateFile string
	esphomeMigrateTemplateData string

	esphomeMigrateCreatePlatform string
	esphomeMigrateCreateBoard    string
	esphomeMigrateCreateSSID     string
	esphomeMigrateCreatePSK      string
)

var esphomeMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migration helpers for ESPHome onboarding",
}

var esphomeMigrateTasmotaCmd = &cobra.Command{
	Use:   "tasmota-template",
	Short: "Analyze and convert Tasmota template exports",
}

var esphomeMigrateTasmotaAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a Tasmota template and show conversion hints",
	RunE: func(cmd *cobra.Command, args []string) error {
		textMode := getTextMode()

		template, err := loadTasmotaTemplateInput()
		if err != nil {
			return err
		}

		analysis := esphometasmota.Analyze(template)
		overlay, _ := esphometasmota.BuildOverlay(template)

		output.PrintOutput(map[string]any{
			"analysis": analysis,
			"overlay":  overlay,
		}, textMode, "")
		return nil
	},
}

var esphomeMigrateTasmotaCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an ESPHome config from a Tasmota template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if esphomeMigrateCreatePlatform == "" {
			return fmt.Errorf("--platform is required")
		}
		if esphomeMigrateCreateBoard == "" {
			return fmt.Errorf("--board is required")
		}

		textMode := getTextMode()
		template, err := loadTasmotaTemplateInput()
		if err != nil {
			return err
		}

		overlay, analysis := esphometasmota.BuildOverlay(template)

		esClient, err := getESPHomeClient()
		if err != nil {
			return err
		}

		platform, err := esphomeconfig.NormalizeWizardPlatform(esphomeMigrateCreatePlatform)
		if err != nil {
			return err
		}

		createResult, err := esClient.CreateConfig(client.ESPHomeCreateRequest{
			Type:     "basic",
			Name:     args[0],
			Platform: platform,
			Board:    esphomeMigrateCreateBoard,
			SSID:     esphomeMigrateCreateSSID,
			PSK:      esphomeMigrateCreatePSK,
		})
		if err != nil {
			return err
		}

		configuration := createResult.Configuration
		original, err := esClient.ReadConfig(configuration)
		if err != nil {
			return err
		}

		updated, err := esphomeconfig.ApplyPatch(original, overlay, nil)
		if err != nil {
			return err
		}

		if updated != original {
			if err := esClient.WriteConfig(configuration, updated); err != nil {
				return err
			}
			if _, err := esClient.GetJSONConfig(configuration); err != nil {
				return withESPHomeRollbackError(err, rollbackESPHomeConfig(esClient, configuration, original))
			}
		}

		output.PrintOutput(map[string]any{
			"configuration": configuration,
			"analysis":      analysis,
			"overlay":       overlay,
		}, textMode, "")
		return nil
	},
}

func init() {
	esphomeCmd.AddCommand(esphomeMigrateCmd)
	esphomeMigrateCmd.AddCommand(esphomeMigrateTasmotaCmd)
	esphomeMigrateTasmotaCmd.AddCommand(esphomeMigrateTasmotaAnalyzeCmd)
	esphomeMigrateTasmotaCmd.AddCommand(esphomeMigrateTasmotaCreateCmd)

	for _, migrateCmd := range []*cobra.Command{esphomeMigrateTasmotaAnalyzeCmd, esphomeMigrateTasmotaCreateCmd} {
		migrateCmd.Flags().StringVarP(&esphomeMigrateTemplateFile, "file", "f", "", "Path to Tasmota template JSON file")
		migrateCmd.Flags().StringVarP(&esphomeMigrateTemplateData, "data", "d", "", "Inline Tasmota template JSON")
	}

	esphomeMigrateTasmotaCreateCmd.Flags().StringVar(&esphomeMigrateCreatePlatform, "platform", "", "ESPHome platform for scaffold (esp32, esp8266, rp2040, ...)")
	esphomeMigrateTasmotaCreateCmd.Flags().StringVar(&esphomeMigrateCreateBoard, "board", "", "ESPHome board ID for scaffold")
	esphomeMigrateTasmotaCreateCmd.Flags().StringVar(&esphomeMigrateCreateSSID, "ssid", "", "WiFi SSID for scaffold")
	esphomeMigrateTasmotaCreateCmd.Flags().StringVar(&esphomeMigrateCreatePSK, "psk", "", "WiFi password for scaffold")
}

func loadTasmotaTemplateInput() (*esphometasmota.Template, error) {
	raw := esphomeMigrateTemplateData
	if strings.TrimSpace(esphomeMigrateTemplateFile) != "" {
		data, err := os.ReadFile(esphomeMigrateTemplateFile)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", esphomeMigrateTemplateFile, err)
		}
		raw = string(data)
	}

	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("provide a tasmota template with --file or --data")
	}

	return esphometasmota.ParseTemplate(raw)
}

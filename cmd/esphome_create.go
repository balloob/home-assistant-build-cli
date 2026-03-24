package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/esphomecatalog"
	"github.com/home-assistant/hab/internal/esphomeconfig"
	"github.com/home-assistant/hab/internal/esphomepreset"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	esphomeCreateType     string
	esphomeCreatePlatform string
	esphomeCreateBoard    string
	esphomeCreateSSID     string
	esphomeCreatePSK      string
	esphomeCreateFile     string
	esphomeCreatePreset   string

	esphomeCreateListPresets bool
	esphomeCreatePresetHelp  string
	esphomeCreateRelayPin    string
	esphomeCreateButtonPin   string
	esphomeCreatePWMPin      string
	esphomeCreateI2CSDA      string
	esphomeCreateI2CSCL      string

	esphomeCreateCatalog    string
	esphomeCreateCatalogRef string
)

var esphomeCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new ESPHome device configuration",
	Long: `Create a new ESPHome configuration from scratch using the dashboard wizard.

Use --type basic to scaffold a working config with wifi/api/ota defaults,
--type empty for a blank file, or --type upload to seed from an existing YAML file.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runESPHomeCreate,
}

func init() {
	esphomeCmd.AddCommand(esphomeCreateCmd)
	esphomeCreateCmd.Flags().StringVar(&esphomeCreateType, "type", "basic", "Wizard type: basic, empty, or upload")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreatePlatform, "platform", "", "ESPHome platform for basic scaffolds (esp32, esp8266, rp2040, bk72xx, ln882x, rtl87xx)")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreateBoard, "board", "", "Board ID for basic scaffolds")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreateSSID, "ssid", "", "WiFi SSID for basic scaffolds")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreatePSK, "psk", "", "WiFi password for basic scaffolds")
	esphomeCreateCmd.Flags().StringVarP(&esphomeCreateFile, "file", "f", "", "Local YAML file to upload when --type upload is used")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreatePreset, "preset", "", "Starter preset to apply after create (relay, plug, light, button, sensor, display-base)")
	esphomeCreateCmd.Flags().BoolVar(&esphomeCreateListPresets, "list-presets", false, "List available starter presets and exit")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreatePresetHelp, "preset-help", "", "Show detailed help for a preset and exit")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreateRelayPin, "relay-pin", "", "Relay GPIO pin for relay/plug presets")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreateButtonPin, "button-pin", "", "Button GPIO pin for button preset")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreatePWMPin, "pwm-pin", "", "PWM GPIO pin for light preset")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreateI2CSDA, "i2c-sda", "", "I2C SDA GPIO pin for sensor/display presets")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreateI2CSCL, "i2c-scl", "", "I2C SCL GPIO pin for sensor/display presets")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreateCatalog, "catalog", "", "Catalog device slug to scaffold from esphome-devices")
	esphomeCreateCmd.Flags().StringVar(&esphomeCreateCatalogRef, "catalog-ref", "main", "Catalog git ref or branch")
}

func runESPHomeCreate(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()
	if esphomeCreateListPresets {
		output.PrintOutput(esphomepreset.List(), textMode, "")
		return nil
	}

	if esphomeCreatePresetHelp != "" {
		preset, ok := esphomepreset.Get(esphomeCreatePresetHelp)
		if !ok {
			return fmt.Errorf("unknown preset %q", esphomeCreatePresetHelp)
		}
		output.PrintOutput(preset, textMode, "")
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("device name is required")
	}
	deviceName := args[0]

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	data := map[string]any{}
	configuration := ""
	if esphomeCreateCatalog != "" {
		configuration, data, err = createESPHomeFromCatalog(esClient, deviceName)
		if err != nil {
			return err
		}
	} else {
		req, reqErr := buildESPHomeCreateRequest(deviceName)
		if reqErr != nil {
			return reqErr
		}

		result, createErr := esClient.CreateConfig(req)
		if createErr != nil {
			return createErr
		}
		configuration = result.Configuration
		data["create_type"] = req.Type
	}

	data["configuration"] = configuration

	original, err := esClient.ReadConfig(configuration)
	if err == nil {
		for key, value := range esphomeconfig.ExtractCreateDetails(original) {
			data[key] = value
		}
	}

	if strings.TrimSpace(esphomeCreatePreset) != "" {
		overlay, preset, presetErr := esphomepreset.BuildOverlay(esphomeCreatePreset, esphomepreset.Options{
			RelayPin:  esphomeCreateRelayPin,
			ButtonPin: esphomeCreateButtonPin,
			PWMPin:    esphomeCreatePWMPin,
			I2CSDA:    esphomeCreateI2CSDA,
			I2CSCL:    esphomeCreateI2CSCL,
		})
		if presetErr != nil {
			return presetErr
		}

		if original == "" {
			readContent, readErr := esClient.ReadConfig(configuration)
			if readErr != nil {
				return readErr
			}
			original = readContent
		}

		updated, patchErr := esphomeconfig.ApplyPatch(original, overlay, nil)
		if patchErr != nil {
			return patchErr
		}

		if updated != original {
			if err := esClient.WriteConfig(configuration, updated); err != nil {
				return err
			}

			if _, err := esClient.GetJSONConfig(configuration); err != nil {
				return withESPHomeRollbackError(err, rollbackESPHomeConfig(esClient, configuration, original))
			}

			for key, value := range esphomeconfig.ExtractCreateDetails(updated) {
				data[key] = value
			}
		}

		data["preset"] = preset.ID
		data["preset_summary"] = preset.Summary
	}

	if _, ok := data["content"]; !ok {
		if content, readErr := esClient.ReadConfig(configuration); readErr == nil {
			for key, value := range esphomeconfig.ExtractCreateDetails(content) {
				data[key] = value
			}
		}
	}

	output.PrintOutput(data, textMode, "")
	return nil
}

func createESPHomeFromCatalog(esClient client.ESPHomeAPI, deviceName string) (string, map[string]any, error) {
	catalogClient := esphomecatalog.NewClient(esphomeCreateCatalogRef)
	device, err := catalogClient.GetDevice(context.Background(), esphomeCreateCatalog)
	if err != nil {
		return "", nil, err
	}

	data := map[string]any{
		"catalog_slug":       device.Slug,
		"catalog_title":      device.Title,
		"catalog_board":      device.Board,
		"catalog_type":       device.DeviceType,
		"catalog_difficulty": device.Difficulty,
		"catalog_ref":        esphomeCreateCatalogRef,
		"catalog_source":     device.ContentSource,
		"catalog_yaml_path":  device.YAMLPath,
	}

	var importErr error
	if strings.TrimSpace(device.PackageImportURL) != "" {
		importResult, err := esClient.ImportConfig(client.ESPHomeImportRequest{
			Name:             deviceName,
			FriendlyName:     device.Title,
			ProjectName:      device.ProjectName,
			PackageImportURL: device.PackageImportURL,
			Encryption:       false,
		})
		if err != nil {
			importErr = err
			data["package_import_url"] = device.PackageImportURL
			data["import_fallback"] = true
			data["import_error"] = err.Error()
		} else {
			data["create_type"] = "catalog_import"
			data["project_name"] = device.ProjectName
			data["package_import_url"] = device.PackageImportURL
			return importResult.Configuration, data, nil
		}
	}

	if strings.TrimSpace(device.YAMLContent) == "" {
		if importErr != nil {
			return "", nil, importErr
		}
		return "", nil, fmt.Errorf("catalog device %q has no import URL or YAML content", device.Slug)
	}

	result, err := esClient.CreateConfig(client.ESPHomeCreateRequest{
		Type:        "upload",
		Name:        deviceName,
		FileContent: []byte(device.YAMLContent),
	})
	if err != nil {
		return "", nil, err
	}

	if err := alignCatalogConfigurationName(esClient, result.Configuration); err != nil {
		return "", nil, err
	}

	data["create_type"] = "catalog_upload"
	return result.Configuration, data, nil
}

func alignCatalogConfigurationName(esClient client.ESPHomeAPI, configuration string) error {
	original, err := esClient.ReadConfig(configuration)
	if err != nil {
		return err
	}

	deviceName := strings.TrimSuffix(configuration, ".yaml")
	updated, err := esphomeconfig.SetDeviceName(original, deviceName)
	if err != nil {
		return err
	}
	if updated == original {
		return nil
	}

	if err := esClient.WriteConfig(configuration, updated); err != nil {
		return err
	}
	if _, err := esClient.GetJSONConfig(configuration); err != nil {
		return withESPHomeRollbackError(err, rollbackESPHomeConfig(esClient, configuration, original))
	}

	return nil
}

func buildESPHomeCreateRequest(name string) (client.ESPHomeCreateRequest, error) {
	req := client.ESPHomeCreateRequest{
		Type: esphomeCreateType,
		Name: name,
	}

	switch esphomeCreateType {
	case "basic":
		if esphomeCreateBoard == "" {
			return client.ESPHomeCreateRequest{}, fmt.Errorf("--board is required for --type basic")
		}
		if esphomeCreatePlatform == "" {
			return client.ESPHomeCreateRequest{}, fmt.Errorf("--platform is required for --type basic")
		}
		platform, err := normalizeESPHomeWizardPlatform(esphomeCreatePlatform)
		if err != nil {
			return client.ESPHomeCreateRequest{}, err
		}
		req.Platform = platform
		req.Board = esphomeCreateBoard
		req.SSID = esphomeCreateSSID
		req.PSK = esphomeCreatePSK
	case "empty":
		return req, nil
	case "upload":
		if esphomeCreateFile == "" {
			return client.ESPHomeCreateRequest{}, fmt.Errorf("--file is required for --type upload")
		}
		data, err := os.ReadFile(esphomeCreateFile)
		if err != nil {
			return client.ESPHomeCreateRequest{}, fmt.Errorf("failed to read %q: %w", esphomeCreateFile, err)
		}
		req.FileContent = data
	default:
		return client.ESPHomeCreateRequest{}, fmt.Errorf("unsupported create type %q", esphomeCreateType)
	}

	return req, nil
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	esphomeImportFriendlyName string
	esphomeImportProjectName  string
	esphomeImportPackageURL   string
	esphomeImportEncryption   bool
)

var esphomeImportCmd = &cobra.Command{
	Use:   "import <name>",
	Short: "Import an ESPHome package-based device profile",
	Long: `Import an ESPHome device from a package URL, which is useful for device
templates and catalog-backed scaffolding flows.`,
	Args: cobra.ExactArgs(1),
	RunE: runESPHomeImport,
}

func init() {
	esphomeCmd.AddCommand(esphomeImportCmd)
	esphomeImportCmd.Flags().StringVar(&esphomeImportFriendlyName, "friendly-name", "", "Optional friendly name for the imported device")
	esphomeImportCmd.Flags().StringVar(&esphomeImportProjectName, "project-name", "", "Project name reported by the imported package")
	esphomeImportCmd.Flags().StringVar(&esphomeImportPackageURL, "package-url", "", "Package import URL to fetch")
	esphomeImportCmd.Flags().BoolVar(&esphomeImportEncryption, "encryption", false, "Enable API encryption during import when supported")
	_ = esphomeImportCmd.MarkFlagRequired("project-name")
	_ = esphomeImportCmd.MarkFlagRequired("package-url")
}

func runESPHomeImport(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()
	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	result, err := esClient.ImportConfig(client.ESPHomeImportRequest{
		Name:             args[0],
		FriendlyName:     esphomeImportFriendlyName,
		ProjectName:      esphomeImportProjectName,
		PackageImportURL: esphomeImportPackageURL,
		Encryption:       esphomeImportEncryption,
	})
	if err != nil {
		return err
	}

	data := map[string]any{
		"configuration": result.Configuration,
		"project_name":  esphomeImportProjectName,
		"package_url":   esphomeImportPackageURL,
	}
	output.PrintOutput(data, textMode, fmt.Sprintf("Imported ESPHome device %s.", args[0]))
	return nil
}

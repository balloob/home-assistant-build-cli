package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	esphomeConfigPatchInput    InputFlags
	esphomeConfigPatchSet      []string
	esphomeConfigPatchValidate bool
	esphomeConfigPatchRollback bool
)

var esphomeConfigPatchCmd = &cobra.Command{
	Use:   "config-patch [configuration]",
	Short: "Patch an ESPHome YAML configuration safely",
	Long: `Read the current ESPHome YAML, apply a structured patch, write the updated
configuration back to the dashboard, and optionally validate the result with
automatic rollback on failure.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runESPHomeConfigPatch,
}

func init() {
	esphomeCmd.AddCommand(esphomeConfigPatchCmd)
	esphomeConfigPatchInput.Register(esphomeConfigPatchCmd)
	esphomeConfigPatchCmd.Flags().StringArrayVar(&esphomeConfigPatchSet, "set", nil, "Set a dotted YAML path to a YAML/JSON value (repeatable)")
	esphomeConfigPatchCmd.Flags().BoolVar(&esphomeConfigPatchValidate, "validate", true, "Validate the updated configuration after writing it")
	esphomeConfigPatchCmd.Flags().BoolVar(&esphomeConfigPatchRollback, "rollback-on-fail", true, "Restore the previous configuration if validation fails")
}

func runESPHomeConfigPatch(cmd *cobra.Command, args []string) error {
	configuration, err := resolveESPHomeConfiguration(args, 0)
	if err != nil {
		return err
	}
	if esphomeConfigPatchInput.Data == "" && esphomeConfigPatchInput.File == "" && len(esphomeConfigPatchSet) == 0 {
		return fmt.Errorf("provide patch data via --data/--file or at least one --set")
	}
	textMode := getTextMode()

	overlay, err := maybeParseESPHomePatchInput(&esphomeConfigPatchInput)
	if err != nil {
		return err
	}

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	original, err := esClient.ReadConfig(configuration)
	if err != nil {
		return err
	}

	updated, err := applyESPHomePatch(original, overlay, esphomeConfigPatchSet)
	if err != nil {
		return err
	}

	written := updated != original
	if written {
		if err := esClient.WriteConfig(configuration, updated); err != nil {
			return err
		}
	}

	if esphomeConfigPatchValidate {
		if _, err := esClient.GetJSONConfig(configuration); err != nil {
			if written && esphomeConfigPatchRollback {
				return withESPHomeRollbackError(err, rollbackESPHomeConfig(esClient, configuration, original))
			}
			return err
		}
	}

	output.PrintOutput(map[string]any{
		"configuration": configuration,
		"validated":     esphomeConfigPatchValidate,
		"written":       written,
	}, textMode, fmt.Sprintf("Configuration %s patched successfully.", configuration))
	return nil
}

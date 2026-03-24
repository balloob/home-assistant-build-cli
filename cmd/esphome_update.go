package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var (
	esphomeUpdateInput    InputFlags
	esphomeUpdateSet      []string
	esphomeUpdateValidate bool
	esphomeUpdateBuild    bool
	esphomeUpdateUpload   bool
	esphomeUpdateRun      bool
	esphomeUpdateRollback bool
	esphomeUpdatePort     string
)

var esphomeUpdateCmd = &cobra.Command{
	Use:   "update [configuration]",
	Short: "Apply a safer ESPHome update workflow",
	Long: `Patch or rewrite an ESPHome configuration, validate it, build firmware, and
optionally upload it in one safer workflow designed for agents and beginners.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runESPHomeUpdate,
}

func init() {
	esphomeCmd.AddCommand(esphomeUpdateCmd)
	esphomeUpdateInput.Register(esphomeUpdateCmd)
	esphomeUpdateCmd.Flags().StringArrayVar(&esphomeUpdateSet, "set", nil, "Set a dotted YAML path to a YAML/JSON value (repeatable)")
	esphomeUpdateCmd.Flags().BoolVar(&esphomeUpdateValidate, "validate", true, "Validate the configuration before running build/upload steps")
	esphomeUpdateCmd.Flags().BoolVar(&esphomeUpdateBuild, "build", true, "Build firmware after validation")
	esphomeUpdateCmd.Flags().BoolVar(&esphomeUpdateUpload, "upload", false, "Upload firmware after a successful build")
	esphomeUpdateCmd.Flags().BoolVar(&esphomeUpdateRun, "run", false, "Run ESPHome's combined build+upload workflow instead of separate steps")
	esphomeUpdateCmd.Flags().BoolVar(&esphomeUpdateRollback, "rollback-on-fail", true, "Restore the previous configuration if any workflow step fails")
	esphomeUpdateCmd.Flags().StringVar(&esphomeUpdatePort, "port", "OTA", "Connection port: OTA (network) or serial port path")
}

func runESPHomeUpdate(cmd *cobra.Command, args []string) error {
	configuration, err := resolveESPHomeConfiguration(args, 0)
	if err != nil {
		return err
	}
	if esphomeUpdateRun && esphomeUpdateUpload {
		return fmt.Errorf("--run cannot be combined with --upload")
	}
	if esphomeUpdateUpload && !esphomeUpdateBuild {
		return fmt.Errorf("--upload requires --build")
	}
	textMode := getTextMode()

	overlay, err := maybeParseESPHomePatchInput(&esphomeUpdateInput)
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

	updated := original
	if overlay != nil || len(esphomeUpdateSet) > 0 {
		updated, err = applyESPHomePatch(original, overlay, esphomeUpdateSet)
		if err != nil {
			return err
		}
	}

	written := updated != original
	if written {
		if err := esClient.WriteConfig(configuration, updated); err != nil {
			return err
		}
	}

	result := map[string]any{
		"configuration": configuration,
		"validated":     false,
		"written":       written,
		"build":         false,
		"upload":        false,
		"run":           false,
	}

	if esphomeUpdateValidate {
		if _, err := esClient.GetJSONConfig(configuration); err != nil {
			if written && esphomeUpdateRollback {
				return withESPHomeRollbackError(err, rollbackESPHomeConfig(esClient, configuration, original))
			}
			return err
		}
		result["validated"] = true
	}

	if esphomeUpdateRun {
		events, err := runESPHomeWorkflowStream(esClient, "/run", map[string]any{
			"type":          "spawn",
			"configuration": configuration,
			"port":          esphomeUpdatePort,
		}, textMode)
		if err != nil {
			if written && esphomeUpdateRollback {
				return withESPHomeRollbackError(err, rollbackESPHomeConfig(esClient, configuration, original))
			}
			return err
		}
		result["run"] = true
		if !textMode {
			result["run_events"] = events
		}
	} else {
		if esphomeUpdateBuild {
			events, err := runESPHomeWorkflowStream(esClient, "/compile", map[string]any{
				"type":          "spawn",
				"configuration": configuration,
				"only_generate": false,
			}, textMode)
			if err != nil {
				if written && esphomeUpdateRollback {
					return withESPHomeRollbackError(err, rollbackESPHomeConfig(esClient, configuration, original))
				}
				return err
			}
			result["build"] = true
			if !textMode {
				result["build_events"] = events
			}
		}

		if esphomeUpdateUpload {
			events, err := runESPHomeWorkflowStream(esClient, "/upload", map[string]any{
				"type":          "spawn",
				"configuration": configuration,
				"port":          esphomeUpdatePort,
			}, textMode)
			if err != nil {
				if written && esphomeUpdateRollback {
					return withESPHomeRollbackError(err, rollbackESPHomeConfig(esClient, configuration, original))
				}
				return err
			}
			result["upload"] = true
			if !textMode {
				result["upload_events"] = events
			}
		}
	}

	if textMode {
		output.PrintSuccess(nil, true, fmt.Sprintf("ESPHome update workflow completed for %s.", configuration))
		return nil
	}

	output.PrintOutput(result, false, "")
	return nil
}

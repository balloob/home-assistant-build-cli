package cmd

import (
	"fmt"
	"os"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/esphomeconfig"
)

func parseESPHomeConfig(content string) (map[string]any, error) {
	return esphomeconfig.ParseConfig(content)
}

func applyESPHomePatch(content string, overlay map[string]any, sets []string) (string, error) {
	return esphomeconfig.ApplyPatch(content, overlay, sets)
}

func extractESPHomeCreateDetails(content string) (map[string]any, error) {
	return esphomeconfig.ExtractCreateDetails(content), nil
}

func maybeParseESPHomePatchInput(flags *InputFlags) (map[string]any, error) {
	if flags == nil {
		return nil, nil
	}
	if flags.Data == "" && flags.File == "" {
		return nil, nil
	}
	return flags.Parse()
}

func rollbackESPHomeConfig(esClient client.ESPHomeAPI, configuration, content string) error {
	if err := esClient.WriteConfig(configuration, content); err != nil {
		return fmt.Errorf("rollback config %s: %w", configuration, err)
	}
	return nil
}

func withESPHomeRollbackError(err, rollbackErr error) error {
	if rollbackErr == nil {
		return err
	}
	return fmt.Errorf("%w (rollback failed: %v)", err, rollbackErr)
}

func normalizeESPHomeWizardPlatform(platform string) (string, error) {
	return esphomeconfig.NormalizeWizardPlatform(platform)
}

func runESPHomeWorkflowStream(esClient client.ESPHomeAPI, wsPath string, spawnMsg map[string]any, textMode bool) ([]client.ESPHomeStreamEvent, error) {
	events := make([]client.ESPHomeStreamEvent, 0)
	exitCode, err := esClient.StreamCommand(wsPath, spawnMsg, func(event client.ESPHomeStreamEvent) {
		event.Data = decodeESPHomeAnsi(event.Data)
		events = append(events, event)
		if textMode && event.Event == "line" {
			_, _ = os.Stdout.WriteString(event.Data)
		}
	})
	if err != nil {
		return events, err
	}
	if exitCode != 0 {
		return events, fmt.Errorf("process exited with code %d", exitCode)
	}
	return events, nil
}

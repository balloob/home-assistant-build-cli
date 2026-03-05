package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var esphomeCmd = &cobra.Command{
	Use:     "esphome",
	Short:   "Manage ESPHome devices",
	Long:    `List, build, validate, and manage ESPHome devices via the ESPHome Dashboard.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(esphomeCmd)
}

// getESPHomeClient is the shared helper used by all esphome subcommands.
// It resolves the dashboard URL (via HAB_ESPHOME_URL or ingress auto-discovery)
// and returns a configured ESPHomeClient implementing client.ESPHomeAPI.
func getESPHomeClient() (client.ESPHomeAPI, error) {
	creds, err := getCredentials()
	if err != nil || creds == nil {
		return nil, err
	}

	esphomeURL := os.Getenv("HAB_ESPHOME_URL")
	esphomeSession := os.Getenv("HAB_ESPHOME_SESSION")

	// Allow overriding the Bearer token for ESPHome specifically.
	// This is needed when calling through HA Core's ingress proxy from
	// an addon container, where the Supervisor token won't work and a
	// long-lived access token is required instead.
	token := creds.AccessToken
	if t := os.Getenv("HAB_ESPHOME_TOKEN"); t != "" {
		token = t
	} else if t := os.Getenv("HA_ACCESS_TOKEN"); t != "" {
		token = t
	}

	return client.GetESPHomeClient(esphomeURL, esphomeSession, creds.URL, token)
}

// decodeESPHomeAnsi converts literal ESPHome escape sequences (e.g. the
// string \033[32m) into real ANSI escape bytes so terminals render colors.
func decodeESPHomeAnsi(s string) string {
	return strings.ReplaceAll(s, `\033`, "\033")
}

// streamToOutput is a shared helper that handles streaming ESPHome WebSocket
// commands (build, logs, validate, upload, run) with proper text/JSON output.
func streamToOutput(esClient client.ESPHomeAPI, wsPath string, spawnMsg map[string]interface{}, textMode bool) error {
	if textMode {
		exitCode, err := esClient.StreamCommand(wsPath, spawnMsg, func(event client.ESPHomeStreamEvent) {
			switch event.Event {
			case "line":
				fmt.Print(decodeESPHomeAnsi(event.Data))
			case "exit":
				// handled below
			}
		})
		if err != nil {
			return err
		}
		if exitCode != 0 {
			return fmt.Errorf("process exited with code %d", exitCode)
		}
		return nil
	}

	// JSON mode: emit NDJSON - one JSON object per event
	exitCode, err := esClient.StreamCommand(wsPath, spawnMsg, func(event client.ESPHomeStreamEvent) {
		// Decode ANSI escapes so consumers get real escape bytes
		event.Data = decodeESPHomeAnsi(event.Data)
		b, _ := json.Marshal(event)
		fmt.Println(string(b))
	})
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("process exited with code %d", exitCode)
	}
	return nil
}

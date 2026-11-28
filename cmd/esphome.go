package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/config"
	"github.com/spf13/cobra"
)

var esphomeDevice string

var esphomeCmd = &cobra.Command{
	Use:   "esphome",
	Short: "Manage ESPHome devices",
	Long: `Create, validate, build, patch, and manage ESPHome devices via the ESPHome Device Builder.

For workflow guidance, run 'hab guide esphome'.`,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(esphomeCmd)
	esphomeCmd.PersistentFlags().StringVar(&esphomeDevice, "device", "", "ESPHome configuration filename to use (falls back to saved context)")
}

// getESPHomeClient is the shared helper used by all esphome subcommands.
// It resolves the Device Builder URL (via HAB_ESPHOME_URL or ingress auto-discovery)
// and returns a configured ESPHomeClient implementing client.ESPHomeAPI.
func getESPHomeClient() (client.ESPHomeAPI, error) {
	creds, err := getCredentials()
	if err != nil || creds == nil {
		return nil, err
	}

	// These variables describe one instance's dashboard; applying them to a
	// named context would mix instances or send its token to another host.
	if manager := getAuthManager(); manager.Context != config.DefaultContext && creds.Source == "credentials_file" {
		for _, name := range []string{"HAB_ESPHOME_URL", "HAB_ESPHOME_TOKEN", "HA_ACCESS_TOKEN"} {
			if os.Getenv(name) != "" {
				return nil, fmt.Errorf("conflicting ESPHome settings: context %q is selected, but %s is set; unset it to use the ESPHome dashboard of that context", manager.Context, name)
			}
		}
	}

	esphomeURL := os.Getenv("HAB_ESPHOME_URL")
	esphomeSession := os.Getenv("HAB_ESPHOME_SESSION")

	// Allow overriding the Bearer token for ESPHome specifically.
	// This is needed when calling through HA Core's ingress proxy from
	// an app container, where the Supervisor token won't work and a
	// long-lived access token is required instead.
	token := creds.AccessToken
	if t := os.Getenv("HAB_ESPHOME_TOKEN"); t != "" {
		token = t
		noteResolution("esphome_token_source", "HAB_ESPHOME_TOKEN")
	} else if t := os.Getenv("HA_ACCESS_TOKEN"); t != "" {
		token = t
		noteResolution("esphome_token_source", "HA_ACCESS_TOKEN")
	} else {
		noteResolution("esphome_token_source", "home_assistant_credentials")
	}
	if esphomeURL != "" {
		noteResolution("esphome_url_source", "HAB_ESPHOME_URL")
	} else {
		noteResolution("esphome_url_source", "auto_discovery")
	}
	noteTransport("esphome")

	return client.GetESPHomeClient(esphomeURL, esphomeSession, creds.URL, token)
}

// decodeESPHomeAnsi converts literal ESPHome escape sequences (e.g. the
// string \033[32m) into real ANSI escape bytes so terminals render colors.
func decodeESPHomeAnsi(s string) string {
	return strings.ReplaceAll(s, `\033`, "\033")
}

// streamToOutput is a shared helper that handles streaming ESPHome
// operations (build, logs, validate, upload, run) with proper text/JSON output.
func streamToOutput(esClient client.ESPHomeAPI, operation, configuration, port string, textMode bool) error {
	if textMode {
		exitCode, err := esClient.StreamCommand(operation, configuration, port, func(event client.ESPHomeStreamEvent) {
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
	exitCode, err := esClient.StreamCommand(operation, configuration, port, func(event client.ESPHomeStreamEvent) {
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

func resolveESPHomeConfiguration(args []string, index int) (string, error) {
	if esphomeDevice != "" && len(args) > index && esphomeDevice != args[index] {
		return "", fmt.Errorf("conflicting configuration values: positional %q does not match --device value %q", args[index], esphomeDevice)
	}
	if esphomeDevice != "" {
		noteResolution("esphome_configuration", "flag")
		return esphomeDevice, nil
	}
	if len(args) > index {
		noteResolution("esphome_configuration", "positional")
		return args[index], nil
	}

	dir, err := getAuthManager().ContextDir()
	if err != nil {
		return "", err
	}
	ctx, err := config.LoadESPHomeContext(dir)
	if err != nil {
		return "", fmt.Errorf("load ESPHome context: %w", err)
	}
	if ctx != nil && ctx.Configuration != "" {
		noteResolution("esphome_configuration", "saved_context")
		return ctx.Configuration, nil
	}

	return "", fmt.Errorf("configuration is required (argument, --device, or 'hab esphome context use <configuration>')")
}

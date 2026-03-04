package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/input"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getWSClient creates an authenticated, connected WebSocket client.
// Caller must defer ws.Close() after a successful return.
func getWSClient() (client.WebSocketAPI, error) {
	configDir := viper.GetString("config")
	manager := auth.NewManager(configDir)
	creds, err := manager.GetCredentials()
	if err != nil || creds == nil {
		return nil, err
	}

	ws := client.NewWebSocketClient(creds.URL, creds.AccessToken)
	if err := ws.Connect(); err != nil {
		return nil, err
	}
	return ws, nil
}

// getRESTClient creates an authenticated REST client.
func getRESTClient() (client.RestAPI, error) {
	configDir := viper.GetString("config")
	manager := auth.NewManager(configDir)
	return manager.GetRestClient()
}

// getCredentials returns the current authentication credentials.
func getCredentials() (*auth.Credentials, error) {
	configDir := viper.GetString("config")
	manager := auth.NewManager(configDir)
	return manager.GetCredentials()
}

// getTextMode returns whether text output mode is enabled.
func getTextMode() bool {
	return viper.GetBool("text")
}

// resolveArg resolves a value from either a flag variable or a positional argument.
// It checks the flag value first; if empty, falls back to args[index].
// Returns an error if no value is found.
func resolveArg(flagVal string, args []string, index int, name string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	if len(args) > index {
		return args[index], nil
	}
	return "", fmt.Errorf("%s is required", name)
}

// confirmAction prompts the user for confirmation unless force or textMode is set.
// Returns true if the action should proceed, false if the user declined.
func confirmAction(force, textMode bool, prompt string) bool {
	if force || textMode {
		return true
	}
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// InputFlags holds the standard --data, --file, --format flag values
// used by commands that accept JSON/YAML configuration input.
type InputFlags struct {
	Data   string
	File   string
	Format string
}

// Register adds the --data/-d, --file/-f, and --format flags to a cobra command.
func (f *InputFlags) Register(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.Data, "data", "d", "", "Configuration data as JSON or YAML string")
	cmd.Flags().StringVarP(&f.File, "file", "f", "", "Path to configuration file")
	cmd.Flags().StringVar(&f.Format, "format", "", "Input format: json or yaml (auto-detected if not specified)")
}

// Parse parses the input data from the flag values, returning a map.
func (f *InputFlags) Parse() (map[string]interface{}, error) {
	return input.ParseInput(f.Data, f.File, f.Format)
}

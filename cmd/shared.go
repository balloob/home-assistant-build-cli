package cmd

import (
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
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

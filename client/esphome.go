package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// ESPHomeClient communicates with the ESPHome Dashboard API.
// The dashboard may be accessed directly or via the HA Supervisor ingress proxy.
type ESPHomeClient struct {
	BaseURL        string
	Token          string // HA bearer token (used for ingress proxy auth)
	IngressSession string // Ingress session cookie value (for HA ingress proxy)
	Timeout        time.Duration
	VerifySSL      bool
	client         *resty.Client
	clientOnce     sync.Once
}

// ESPHomeDevice represents a configured device from the ESPHome dashboard.
type ESPHomeDevice struct {
	Name               string   `json:"name"`
	FriendlyName       string   `json:"friendly_name"`
	Configuration      string   `json:"configuration"`
	LoadedIntegrations []string `json:"loaded_integrations"`
	DeployedVersion    string   `json:"deployed_version"`
	CurrentVersion     string   `json:"current_version"`
	Path               string   `json:"path"`
	Comment            *string  `json:"comment"`
	Address            string   `json:"address"`
	WebPort            *int     `json:"web_port"`
	TargetPlatform     string   `json:"target_platform"`
}

// ESPHomeImportableDevice represents a discovered but not yet imported device.
type ESPHomeImportableDevice struct {
	Name             string `json:"name"`
	FriendlyName     string `json:"friendly_name"`
	PackageImportURL string `json:"package_import_url"`
	ProjectName      string `json:"project_name"`
	ProjectVersion   string `json:"project_version"`
	Network          string `json:"network"`
	Ignored          bool   `json:"ignored"`
}

// ESPHomeDeviceList is the response from GET /devices.
type ESPHomeDeviceList struct {
	Configured []ESPHomeDevice           `json:"configured"`
	Importable []ESPHomeImportableDevice `json:"importable"`
}

// ESPHomeStreamEvent represents a line or exit event from a streaming WebSocket command.
type ESPHomeStreamEvent struct {
	Event string `json:"event"` // "line" or "exit"
	Data  string `json:"data,omitempty"`
	Code  *int   `json:"code,omitempty"`
}

// NewESPHomeClient creates a new ESPHome Dashboard client.
// baseURL should be the full ingress URL or direct dashboard URL.
// token is the HA bearer token (used when proxying through ingress).
func NewESPHomeClient(baseURL, token string) *ESPHomeClient {
	return &ESPHomeClient{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		Token:     token,
		Timeout:   DefaultTimeout,
		VerifySSL: true,
	}
}

func (c *ESPHomeClient) getClient() *resty.Client {
	c.clientOnce.Do(func() {
		c.client = resty.New()
		c.client.SetTimeout(c.Timeout)
		c.client.SetBaseURL(c.BaseURL)
		if c.Token != "" {
			c.client.SetHeader("Authorization", "Bearer "+c.Token)
		}
		if c.IngressSession != "" {
			c.client.SetCookie(&http.Cookie{
				Name:  "ingress_session",
				Value: c.IngressSession,
			})
		}

		if cfg := tlsConfig(c.VerifySSL); cfg != nil {
			c.client.SetTLSClientConfig(cfg)
		}

		c.client.OnAfterResponse(func(client *resty.Client, resp *resty.Response) error {
			log.WithFields(log.Fields{
				"status": resp.StatusCode(),
				"time":   resp.Time(),
				"url":    resp.Request.URL,
				"method": resp.Request.Method,
			}).Debug("ESPHome REST response")
			return nil
		})
	})
	return c.client
}

// doGet performs a GET request, checks for a 200 status, and JSON-unmarshals
// the response body into dest. operation names the action for error messages.
func (c *ESPHomeClient) doGet(path, operation string, dest interface{}) error {
	resp, err := c.getClient().R().Get(path)
	if err != nil {
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to %s: %s", operation, err)}
	}
	if resp.StatusCode() != http.StatusOK {
		return &APIError{
			Code:    ErrCodeAPIError,
			Message: fmt.Sprintf("ESPHome dashboard returned status %d: %s", resp.StatusCode(), string(resp.Body())),
		}
	}
	if err := json.Unmarshal(resp.Body(), dest); err != nil {
		return &APIError{Code: ErrCodeAPIError, Message: fmt.Sprintf("failed to parse %s response: %s", operation, err)}
	}
	return nil
}

// GetDevices returns the list of configured and importable ESPHome devices.
func (c *ESPHomeClient) GetDevices() (*ESPHomeDeviceList, error) {
	var result ESPHomeDeviceList
	if err := c.doGet("/devices", "list devices", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPing returns online/offline status for all devices.
// Map keys are configuration filenames, values are true (online), false (offline), or nil (unknown).
func (c *ESPHomeClient) GetPing() (map[string]*bool, error) {
	var result map[string]*bool
	if err := c.doGet("/ping", "ping devices", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetVersion returns the ESPHome version string.
func (c *ESPHomeClient) GetVersion() (string, error) {
	var result struct {
		Version string `json:"version"`
	}
	if err := c.doGet("/version", "get version", &result); err != nil {
		return "", err
	}
	return result.Version, nil
}

// ReadConfig reads the YAML configuration for a device.
func (c *ESPHomeClient) ReadConfig(configuration string) (string, error) {
	resp, err := c.getClient().R().
		SetQueryParam("configuration", configuration).
		Get("/edit")
	if err != nil {
		return "", &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to read config: %s", err)}
	}
	if resp.StatusCode() == http.StatusNotFound {
		return "", &APIError{Code: ErrCodeNotFound, Message: fmt.Sprintf("configuration %q not found", configuration)}
	}
	if resp.StatusCode() != http.StatusOK {
		return "", &APIError{Code: ErrCodeAPIError, Message: fmt.Sprintf("ESPHome dashboard returned status %d", resp.StatusCode())}
	}
	return string(resp.Body()), nil
}

// WriteConfig writes YAML configuration for a device.
func (c *ESPHomeClient) WriteConfig(configuration, content string) error {
	resp, err := c.getClient().R().
		SetQueryParam("configuration", configuration).
		SetHeader("Content-Type", "application/yaml").
		SetBody(content).
		Post("/edit")
	if err != nil {
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to write config: %s", err)}
	}
	if resp.StatusCode() != http.StatusOK {
		return &APIError{Code: ErrCodeAPIError, Message: fmt.Sprintf("ESPHome dashboard returned status %d: %s", resp.StatusCode(), string(resp.Body()))}
	}
	return nil
}

// StreamCommand opens a WebSocket to the ESPHome dashboard and runs a streaming
// command (compile, logs, validate, upload, run). It sends the spawn message,
// then calls the callback for each event received until the process exits.
// The callback receives ESPHomeStreamEvent with event "line" (output) or "exit" (done).
// Returns the exit code of the subprocess, or an error if the connection fails.
func (c *ESPHomeClient) StreamCommand(path string, spawnMsg map[string]interface{}, callback func(ESPHomeStreamEvent)) (int, error) {
	wsURL, err := c.buildWSURL(path)
	if err != nil {
		return -1, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to build WebSocket URL: %s", err)}
	}

	log.WithField("url", wsURL).Debug("ESPHome WebSocket connecting")

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	dialer.TLSClientConfig = tlsConfig(c.VerifySSL)

	headers := http.Header{}
	if c.Token != "" {
		headers.Set("Authorization", "Bearer "+c.Token)
	}
	if c.IngressSession != "" {
		headers.Set("Cookie", "ingress_session="+c.IngressSession)
	}

	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		if resp != nil {
			return -1, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("ESPHome WebSocket connection failed (%d): %s", resp.StatusCode, err)}
		}
		return -1, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("ESPHome WebSocket connection failed: %s", err)}
	}
	defer conn.Close()

	// Send the spawn message
	if spawnMsg["type"] == nil {
		spawnMsg["type"] = "spawn"
	}

	log.WithField("msg", spawnMsg).Debug("Sending spawn message")

	if err := conn.WriteJSON(spawnMsg); err != nil {
		return -1, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to send spawn message: %s", err)}
	}

	// Read events until exit
	exitCode := -1
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				break
			}
			// Connection closed by server after exit event is normal
			if exitCode >= 0 {
				break
			}
			return exitCode, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("WebSocket read error: %s", err)}
		}

		var event ESPHomeStreamEvent
		if err := json.Unmarshal(message, &event); err != nil {
			log.WithField("raw", string(message)).Debug("Unparseable ESPHome WS message")
			continue
		}

		callback(event)

		if event.Event == "exit" {
			if event.Code != nil {
				exitCode = *event.Code
			} else {
				exitCode = 0
			}
			break
		}
	}

	return exitCode, nil
}

// buildWSURL converts the REST base URL to a WebSocket URL and appends the path.
func (c *ESPHomeClient) buildWSURL(path string) (string, error) {
	parsed, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}

	switch parsed.Scheme {
	case "https":
		parsed.Scheme = "wss"
	case "http":
		parsed.Scheme = "ws"
	case "ws", "wss":
		// already websocket
	default:
		parsed.Scheme = "ws"
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/" + strings.TrimLeft(path, "/")
	return parsed.String(), nil
}

// ESPHomeIngressInfo contains the discovered ingress URL and session token.
type ESPHomeIngressInfo struct {
	URL     string
	Session string
}

// DiscoverESPHomeIngress uses the HA WebSocket API to find the ESPHome addon
// and returns its ingress URL and a valid session token.
// This works with long-lived access tokens.
// baseURL is the HA instance URL, token is the HA access token.
func DiscoverESPHomeIngress(baseURL, token string) (*ESPHomeIngressInfo, error) {
	base := strings.TrimRight(baseURL, "/")

	// Use WebSocket API to query Supervisor addon info.
	// The REST /api/hassio/ proxy often rejects long-lived tokens,
	// but the WebSocket supervisor/api command works reliably.
	ws := NewWebSocketClient(base, token)
	if err := ws.Connect(); err != nil {
		return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to connect to HA WebSocket: %s", err)}
	}
	defer ws.Close()

	// Known ESPHome addon slugs: official and community
	slugs := []string{"5c53de3b_esphome", "a0d7b954_esphome"}

	for _, slug := range slugs {
		log.WithField("slug", slug).Debug("Trying ESPHome addon slug via WebSocket")

		result, err := ws.SendCommand("supervisor/api", map[string]interface{}{
			"endpoint": fmt.Sprintf("/addons/%s/info", slug),
			"method":   "get",
		})
		if err != nil {
			log.WithError(err).Debug("Supervisor API call failed for slug")
			continue
		}

		// result should be a map with addon info
		data, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		state, _ := data["state"].(string)
		if state != "started" && state != "" {
			return nil, &APIError{Code: ErrCodeAPIError, Message: fmt.Sprintf("ESPHome addon is installed but not running (state: %s)", state)}
		}

		ingressPath, _ := data["ingress_url"].(string)
		if ingressPath == "" {
			ingressPath, _ = data["ingress_entry"].(string)
		}
		if ingressPath == "" {
			continue
		}

		ingressURL := base + ingressPath
		log.WithField("url", ingressURL).Debug("Discovered ESPHome ingress URL")

		// Create an ingress session so we can authenticate requests through the proxy.
		sessionResult, err := ws.SendCommand("supervisor/api", map[string]interface{}{
			"endpoint": "/ingress/session",
			"method":   "post",
		})
		if err != nil {
			return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to create ingress session: %s", err)}
		}

		sessionData, ok := sessionResult.(map[string]interface{})
		if !ok {
			return nil, &APIError{Code: ErrCodeAPIError, Message: "unexpected ingress session response"}
		}
		session, _ := sessionData["session"].(string)
		if session == "" {
			return nil, &APIError{Code: ErrCodeAPIError, Message: "empty ingress session token returned"}
		}

		log.Debug("Created ingress session successfully")

		return &ESPHomeIngressInfo{
			URL:     ingressURL,
			Session: session,
		}, nil
	}

	return nil, &APIError{Code: ErrCodeNotFound, Message: "ESPHome addon not found; set HAB_ESPHOME_URL to the ESPHome dashboard URL"}
}

// GetESPHomeClient creates a fully configured ESPHomeClient by resolving the URL.
// If esphomeURL is set, it is used directly with the optional esphomeSession.
// Otherwise, auto-discovers via HA and creates an ingress session.
func GetESPHomeClient(esphomeURL, esphomeSession, haBaseURL, haToken string) (*ESPHomeClient, error) {
	if esphomeURL != "" {
		c := NewESPHomeClient(strings.TrimRight(esphomeURL, "/"), haToken)
		c.IngressSession = esphomeSession
		return c, nil
	}

	// Auto-discover via HA ingress
	info, err := DiscoverESPHomeIngress(haBaseURL, haToken)
	if err != nil {
		return nil, err
	}

	client := NewESPHomeClient(info.URL, haToken)
	client.IngressSession = info.Session
	return client, nil
}

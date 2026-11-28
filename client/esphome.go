package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// ESPHomeClient communicates with the ESPHome Device Builder API.
// The Device Builder may be accessed directly or via the HA Supervisor ingress proxy.
type ESPHomeClient struct {
	BaseURL        string
	Token          string // HA bearer token (used for ingress proxy auth)
	IngressSession string // Ingress session cookie value (for HA ingress proxy)
	Timeout        time.Duration
	VerifySSL      bool
	client         *resty.Client
	clientOnce     sync.Once
	// checked is set once /ws answered, so REST-only calls know the server is
	// the ESPHome Device Builder.
	checked bool
}

// ESPHomeDevice represents a configured device on the ESPHome Device Builder.
type ESPHomeDevice struct {
	Name               string   `json:"name"`
	FriendlyName       string   `json:"friendly_name"`
	Configuration      string   `json:"configuration"`
	LoadedIntegrations []string `json:"loaded_integrations"`
	DeployedVersion    string   `json:"deployed_version"`
	CurrentVersion     string   `json:"current_version"`
	Comment            *string  `json:"comment"`
	Address            string   `json:"address"`
	WebPort            *int     `json:"web_port"`
	TargetPlatform     string   `json:"target_platform"`
	State              string   `json:"state"` // online, offline or unknown
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

// ESPHomeDeviceList is the response from devices/list.
type ESPHomeDeviceList struct {
	Configured []ESPHomeDevice           `json:"configured"`
	Importable []ESPHomeImportableDevice `json:"importable"`
}

// ESPHomeBoard represents a selectable board for a given platform.
type ESPHomeBoard struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ESPHomeSerialPort represents a serial port attached to the ESPHome Device Builder.
type ESPHomeSerialPort struct {
	Port string `json:"port"`
	Desc string `json:"desc"`
}

// ESPHomeCreateRequest contains the payload for devices/create.
type ESPHomeCreateRequest struct {
	Name        string
	Board       string // board catalog ID, see GetBoards
	SSID        string
	PSK         string
	FileContent []byte
}

// ESPHomeCreateResponse is returned after a config is created successfully.
type ESPHomeCreateResponse struct {
	Configuration string `json:"configuration"`
	Warning       string `json:"warning,omitempty"`
}

// ESPHomeImportRequest contains the payload for devices/import.
type ESPHomeImportRequest struct {
	Name             string
	FriendlyName     string
	ProjectName      string
	PackageImportURL string
	Encryption       bool
}

// ESPHomeImportResponse is returned after an import succeeds.
type ESPHomeImportResponse struct {
	Configuration string `json:"configuration"`
	Warning       string `json:"warning,omitempty"`
}

// ESPHomeStreamEvent represents a line or exit event from a streaming command.
type ESPHomeStreamEvent struct {
	Event string `json:"event"` // "line" or "exit"
	Data  string `json:"data,omitempty"`
	Code  *int   `json:"code,omitempty"`
}

// NewESPHomeClient creates a new ESPHome Device Builder client.
// baseURL should be the full ingress URL or direct Device Builder URL.
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

// GetInfo returns build and platform metadata for a configuration, combining
// its resolved config with the device list.
func (c *ESPHomeClient) GetInfo(configuration string) (map[string]any, error) {
	parsed, err := c.GetJSONConfig(configuration)
	if err != nil {
		return nil, err
	}

	info := map[string]any{
		"configuration": configuration,
		"source":        "json_config",
	}
	mergeParsedESPHomeInfo(info, parsed)

	devices, devicesErr := c.GetDevices()
	if devicesErr == nil {
		if device, ok := findESPHomeDevice(devices, configuration); ok {
			mergeESPHomeDeviceInfo(info, device)
			info["source"] = "devices+json_config"
		}
	}

	if _, ok := info["esphome_version"]; !ok {
		if version, err := c.GetVersion(); err == nil && version != "" {
			info["esphome_version"] = version
		}
	}

	return info, nil
}

func findESPHomeDevice(devices *ESPHomeDeviceList, configuration string) (ESPHomeDevice, bool) {
	if devices == nil {
		return ESPHomeDevice{}, false
	}

	for _, device := range devices.Configured {
		if device.Configuration == configuration {
			return device, true
		}
	}

	return ESPHomeDevice{}, false
}

func mergeParsedESPHomeInfo(info map[string]any, parsed map[string]any) {
	if esphome, ok := parsed["esphome"].(map[string]any); ok {
		for _, key := range []string{"name", "friendly_name", "comment", "build_path"} {
			if value, exists := esphome[key]; exists {
				info[key] = value
			}
		}
	}

	loadedIntegrations := make([]string, 0, len(parsed))
	for key := range parsed {
		loadedIntegrations = append(loadedIntegrations, key)
	}
	slices.Sort(loadedIntegrations)
	if len(loadedIntegrations) > 0 {
		info["loaded_integrations"] = loadedIntegrations
	}

	if loadedPlatforms := extractESPHomeLoadedPlatforms(parsed); len(loadedPlatforms) > 0 {
		info["loaded_platforms"] = loadedPlatforms
	}

	if corePlatform, espPlatform, framework := extractESPHomePlatformInfo(parsed); corePlatform != "" || espPlatform != "" || framework != "" {
		if corePlatform != "" {
			info["core_platform"] = corePlatform
		}
		if espPlatform != "" {
			info["esp_platform"] = espPlatform
		}
		if framework != "" {
			info["framework"] = framework
		}
	}

	if mdns, ok := parsed["mdns"].(map[string]any); ok {
		if disabled, ok := mdns["disabled"].(bool); ok {
			info["no_mdns"] = disabled
		}
	}

	name, _ := info["name"].(string)
	buildPath, _ := info["build_path"].(string)
	if name != "" && buildPath != "" {
		info["firmware_bin_path"] = strings.TrimRight(buildPath, "/") + "/.pioenvs/" + name + "/firmware.bin"
	}
}

func mergeESPHomeDeviceInfo(info map[string]any, device ESPHomeDevice) {
	if _, ok := info["name"]; !ok && device.Name != "" {
		info["name"] = device.Name
	}
	if _, ok := info["friendly_name"]; !ok && device.FriendlyName != "" {
		info["friendly_name"] = device.FriendlyName
	}
	if _, ok := info["comment"]; !ok && device.Comment != nil {
		info["comment"] = *device.Comment
	}
	if _, ok := info["loaded_integrations"]; !ok && len(device.LoadedIntegrations) > 0 {
		info["loaded_integrations"] = device.LoadedIntegrations
	}
	if _, ok := info["esp_platform"]; !ok && device.TargetPlatform != "" {
		info["esp_platform"] = device.TargetPlatform
	}
	if _, ok := info["core_platform"]; !ok {
		if corePlatform := normalizeESPHomeCorePlatform(device.TargetPlatform); corePlatform != "" {
			info["core_platform"] = corePlatform
		}
	}
	if _, ok := info["esphome_version"]; !ok && device.CurrentVersion != "" {
		info["esphome_version"] = device.CurrentVersion
	}
	if device.Address != "" {
		info["address"] = device.Address
	}
	if device.WebPort != nil {
		info["web_port"] = *device.WebPort
	}
}

func extractESPHomeLoadedPlatforms(parsed map[string]any) []string {
	loaded := make([]string, 0)
	seen := map[string]struct{}{}

	for component, value := range parsed {
		switch typed := value.(type) {
		case map[string]any:
			if platform, ok := typed["platform"].(string); ok && platform != "" {
				entry := component + "/" + platform
				if _, ok := seen[entry]; !ok {
					seen[entry] = struct{}{}
					loaded = append(loaded, entry)
				}
			}
		case []any:
			for _, item := range typed {
				itemMap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				platform, ok := itemMap["platform"].(string)
				if !ok || platform == "" {
					continue
				}
				entry := component + "/" + platform
				if _, ok := seen[entry]; ok {
					continue
				}
				seen[entry] = struct{}{}
				loaded = append(loaded, entry)
			}
		}
	}

	slices.Sort(loaded)
	return loaded
}

func extractESPHomePlatformInfo(parsed map[string]any) (string, string, string) {
	for _, key := range []string{"esp32", "esp8266", "rp2040", "bk72xx", "ln882x", "rtl87xx"} {
		section, ok := parsed[key].(map[string]any)
		if !ok {
			continue
		}

		framework := ""
		if frameworkMap, ok := section["framework"].(map[string]any); ok {
			if value, ok := frameworkMap["type"].(string); ok {
				framework = value
			}
		}

		espPlatform := strings.ToUpper(key)
		if key == "esp32" {
			if variant, ok := section["variant"].(string); ok && variant != "" {
				espPlatform = variant
			}
		}

		return key, espPlatform, framework
	}

	return "", "", ""
}

func normalizeESPHomeCorePlatform(target string) string {
	target = strings.ToLower(strings.TrimSpace(target))
	switch {
	case strings.HasPrefix(target, "esp32"):
		return "esp32"
	case strings.HasPrefix(target, "esp8266"):
		return "esp8266"
	case strings.HasPrefix(target, "rp2040"):
		return "rp2040"
	case strings.HasPrefix(target, "bk72"):
		return "bk72xx"
	case strings.HasPrefix(target, "ln882"):
		return "ln882x"
	case strings.HasPrefix(target, "rtl87"):
		return "rtl87xx"
	default:
		return ""
	}
}

// GetJSONConfig validates a configuration and returns the fully resolved config.
// The Device Builder serves this only over REST (/json-config); its /ws API has
// no resolved-config command.
func (c *ESPHomeClient) GetJSONConfig(configuration string) (map[string]any, error) {
	if !c.checked {
		// The legacy dashboard also serves /json-config; refuse it here too.
		if _, err := c.call("ping", nil, nil); err != nil {
			return nil, err
		}
	}
	resp, err := c.getClient().R().
		SetQueryParam("configuration", configuration).
		Get("/json-config")
	if err != nil {
		return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to validate config: %s", err)}
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, &APIError{Code: ErrCodeNotFound, Message: fmt.Sprintf("configuration %q not found", configuration)}
	}
	if resp.StatusCode() == http.StatusUnprocessableEntity {
		return nil, parseESPHomeValidationError(configuration, string(resp.Body()))
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, parseESPHomeHTTPError(resp.StatusCode(), string(resp.Body()), "validate config")
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, &APIError{Code: ErrCodeAPIError, Message: fmt.Sprintf("failed to parse validation response: %s", err)}
	}
	return result, nil
}

func parseESPHomeHTTPError(statusCode int, body string, operation string) error {
	body = strings.TrimSpace(body)
	message := espHomeHTTPErrorMessage(body)
	if message == "" {
		message = fmt.Sprintf("ESPHome Device Builder returned status %d", statusCode)
	}

	code := ErrCodeAPIError
	switch statusCode {
	case http.StatusNotFound:
		code = ErrCodeNotFound
	case http.StatusBadRequest, http.StatusConflict, http.StatusUnprocessableEntity:
		code = ErrCodeValidationError
	}

	details := map[string]any{
		"operation":   operation,
		"status_code": statusCode,
	}
	if body != "" {
		details["raw"] = body
	}

	return &APIError{Code: code, Message: message, Details: details}
}

func espHomeHTTPErrorMessage(body string) string {
	if body == "" {
		return ""
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err == nil {
		for _, key := range []string{"error", "message"} {
			if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}

	return body
}

func parseESPHomeValidationError(configuration, raw string) error {
	raw = strings.TrimSpace(raw)
	message := summarizeESPHomeValidation(raw)
	details := map[string]any{
		"configuration": configuration,
		"error_type":    classifyESPHomeValidation(raw),
		"likely_fix":    suggestESPHomeValidationFix(raw),
		"raw":           raw,
	}

	if line, column, ok := extractESPHomeLineColumn(raw); ok {
		details["line"] = line
		details["column"] = column
	}
	if component := extractESPHomeComponent(raw); component != "" {
		details["component"] = component
	}

	return &APIError{Code: ErrCodeValidationError, Message: message, Details: details}
}

func summarizeESPHomeValidation(raw string) string {
	if raw == "" {
		return "ESPHome validation failed"
	}

	lines := strings.Split(raw, "\n")
	candidates := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		candidates = append(candidates, trimmed)
	}

	for _, candidate := range candidates {
		lower := strings.ToLower(candidate)
		if strings.Contains(lower, "error") || strings.Contains(lower, "invalid") || strings.Contains(lower, "failed") {
			return candidate
		}
	}
	return candidates[len(candidates)-1]
}

func classifyESPHomeValidation(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "while parsing") || strings.Contains(lower, "could not find expected") || strings.Contains(lower, "mapping values are not allowed") || strings.Contains(lower, "found character that cannot start any token"):
		return "yaml_parse"
	case strings.Contains(lower, "unknown board") || strings.Contains(lower, "board is not"):
		return "board"
	case strings.Contains(lower, "unknown component") || strings.Contains(lower, "platform not found"):
		return "component"
	case strings.Contains(lower, "required") && (strings.Contains(lower, "option") || strings.Contains(lower, "key")):
		return "missing_option"
	default:
		return "validation"
	}
}

func suggestESPHomeValidationFix(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "while parsing") || strings.Contains(lower, "could not find expected") || strings.Contains(lower, "mapping values are not allowed") || strings.Contains(lower, "found character that cannot start any token"):
		return "Check YAML indentation, quoting, and list structure near the reported line."
	case strings.Contains(lower, "unknown board") || strings.Contains(lower, "board is not"):
		return "Verify the board ID and compare it against 'hab esphome boards <platform>'."
	case strings.Contains(lower, "unknown component") || strings.Contains(lower, "platform not found"):
		return "Check the component or platform name and confirm it is supported by your ESPHome version."
	case strings.Contains(lower, "required") && (strings.Contains(lower, "option") || strings.Contains(lower, "key")):
		return "Add the missing required option shown in the validation output."
	default:
		return "Review the reported component and line number, then rerun validation after fixing the YAML."
	}
}

func extractESPHomeComponent(raw string) string {
	patterns := []string{
		`(?m)^\[([a-z0-9_]+)\]$`,
		`(?i)component\s+([a-z0-9_]+)`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(raw)
		if len(match) > 1 {
			return strings.ToLower(match[1])
		}
	}
	return ""
}

func extractESPHomeLineColumn(raw string) (int, int, bool) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)line\s+(\d+),\s*column\s+(\d+)`),
		regexp.MustCompile(`:(\d+):(\d+)`),
	}
	for _, pattern := range patterns {
		match := pattern.FindStringSubmatch(raw)
		if len(match) != 3 {
			continue
		}
		line, lineErr := strconv.Atoi(match[1])
		column, columnErr := strconv.Atoi(match[2])
		if lineErr == nil && columnErr == nil {
			return line, column, true
		}
	}
	return 0, 0, false
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

		result, err := ws.SendCommand("supervisor/api", map[string]any{
			"endpoint": fmt.Sprintf("/addons/%s/info", slug),
			"method":   "get",
		})
		if err != nil {
			log.WithError(err).Debug("Supervisor API call failed for slug")
			continue
		}

		// result should be a map with addon info
		data, ok := result.(map[string]any)
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
		sessionResult, err := ws.SendCommand("supervisor/api", map[string]any{
			"endpoint": "/ingress/session",
			"method":   "post",
		})
		if err != nil {
			return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to create ingress session: %s", err)}
		}

		sessionData, ok := sessionResult.(map[string]any)
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

	return nil, &APIError{Code: ErrCodeNotFound, Message: "ESPHome app not found; set HAB_ESPHOME_URL to the ESPHome Device Builder URL"}
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

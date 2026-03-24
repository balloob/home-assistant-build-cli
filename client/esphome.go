package client

import (
	"encoding/base64"
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

// ESPHomeBoard represents a selectable board for a given platform.
type ESPHomeBoard struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ESPHomeSerialPort represents a serial port entry from dashboard discovery.
type ESPHomeSerialPort struct {
	Port string `json:"port"`
	Desc string `json:"desc"`
}

// ESPHomeCreateRequest contains the payload for the dashboard wizard endpoint.
type ESPHomeCreateRequest struct {
	Type        string
	Name        string
	Platform    string
	Board       string
	SSID        string
	PSK         string
	FileContent []byte
}

// ESPHomeCreateResponse is returned after a config is created successfully.
type ESPHomeCreateResponse struct {
	Configuration string `json:"configuration"`
}

// ESPHomeImportRequest contains the payload for the dashboard import endpoint.
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
func (c *ESPHomeClient) doGet(path, operation string, dest any) error {
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

// doJSONPost performs a JSON POST request and optionally decodes the response.
func (c *ESPHomeClient) doJSONPost(path, operation string, body any, dest any) error {
	resp, err := c.getClient().R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(path)
	if err != nil {
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to %s: %s", operation, err)}
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return parseESPHomeDashboardError(resp.StatusCode(), string(resp.Body()), operation)
	}

	if dest == nil || len(resp.Body()) == 0 {
		return nil
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

// GetBoards returns the board catalog for a platform.
func (c *ESPHomeClient) GetBoards(platform string) ([]ESPHomeBoard, error) {
	var result []struct {
		Items map[string]string `json:"items"`
	}
	if err := c.doGet("/boards/"+strings.ToLower(platform), "get boards", &result); err != nil {
		return nil, err
	}

	boards := make([]ESPHomeBoard, 0)
	for _, group := range result {
		for id, name := range group.Items {
			boards = append(boards, ESPHomeBoard{ID: id, Name: name})
		}
	}

	slices.SortFunc(boards, func(a, b ESPHomeBoard) int {
		if cmp := strings.Compare(a.Name, b.Name); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.ID, b.ID)
	})

	return boards, nil
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

// GetSerialPorts lists available serial ports from the ESPHome dashboard.
func (c *ESPHomeClient) GetSerialPorts() ([]ESPHomeSerialPort, error) {
	var result []ESPHomeSerialPort
	if err := c.doGet("/serial-ports", "list serial ports", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetInfo returns dashboard storage metadata for a configuration.
func (c *ESPHomeClient) GetInfo(configuration string) (map[string]any, error) {
	resp, err := c.getClient().R().
		SetQueryParam("configuration", configuration).
		Get("/info")
	if err != nil {
		return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to get info: %s", err)}
	}
	if resp.StatusCode() == http.StatusNotFound {
		return c.getInfoFallback(configuration)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, parseESPHomeDashboardError(resp.StatusCode(), string(resp.Body()), "get info")
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, &APIError{Code: ErrCodeAPIError, Message: fmt.Sprintf("failed to parse info response: %s", err)}
	}
	return result, nil
}

func (c *ESPHomeClient) getInfoFallback(configuration string) (map[string]any, error) {
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
	if device.Path != "" {
		info["path"] = device.Path
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

// CreateConfig creates a new ESPHome configuration via the dashboard wizard.
func (c *ESPHomeClient) CreateConfig(req ESPHomeCreateRequest) (*ESPHomeCreateResponse, error) {
	body := map[string]any{
		"type": req.Type,
		"name": req.Name,
	}
	if req.Platform != "" {
		body["platform"] = req.Platform
	}
	if req.Board != "" {
		body["board"] = req.Board
	}
	if req.SSID != "" {
		body["ssid"] = req.SSID
	}
	if req.PSK != "" {
		body["psk"] = req.PSK
	}
	if len(req.FileContent) > 0 {
		body["file_content"] = base64.StdEncoding.EncodeToString(req.FileContent)
	}

	var result ESPHomeCreateResponse
	if err := c.doJSONPost("/wizard", "create config", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ImportConfig imports a package-backed ESPHome device configuration.
func (c *ESPHomeClient) ImportConfig(req ESPHomeImportRequest) (*ESPHomeImportResponse, error) {
	body := map[string]any{
		"name":               req.Name,
		"project_name":       req.ProjectName,
		"package_import_url": req.PackageImportURL,
		"encryption":         req.Encryption,
	}
	if req.FriendlyName != "" {
		body["friendly_name"] = req.FriendlyName
	}

	var result ESPHomeImportResponse
	if err := c.doJSONPost("/import", "import config", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetJSONConfig validates a configuration and returns the parsed JSON structure.
func (c *ESPHomeClient) GetJSONConfig(configuration string) (map[string]any, error) {
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
		return nil, parseESPHomeDashboardError(resp.StatusCode(), string(resp.Body()), "validate config")
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, &APIError{Code: ErrCodeAPIError, Message: fmt.Sprintf("failed to parse validation response: %s", err)}
	}
	return result, nil
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

func parseESPHomeDashboardError(statusCode int, body string, operation string) error {
	body = strings.TrimSpace(body)
	message := espHomeDashboardErrorMessage(body)
	if message == "" {
		message = fmt.Sprintf("ESPHome dashboard returned status %d", statusCode)
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

func espHomeDashboardErrorMessage(body string) string {
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

// StreamCommand opens a WebSocket to the ESPHome dashboard and runs a streaming
// command (compile, logs, validate, upload, run). It sends the spawn message,
// then calls the callback for each event received until the process exits.
// The callback receives ESPHomeStreamEvent with event "line" (output) or "exit" (done).
// Returns the exit code of the subprocess, or an error if the connection fails.
func (c *ESPHomeClient) StreamCommand(path string, spawnMsg map[string]any, callback func(ESPHomeStreamEvent)) (int, error) {
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

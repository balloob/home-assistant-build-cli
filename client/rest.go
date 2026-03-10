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
	log "github.com/sirupsen/logrus"
)

const (
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 * time.Second
)

// RestClient is an HTTP client for the Home Assistant REST API.
// After initialization, concurrent Get/Post/Put/Delete calls are safe.
type RestClient struct {
	BaseURL   string
	Token     string
	Timeout   time.Duration
	VerifySSL bool
	client    *resty.Client
	clientMu  sync.Once
}

// NewRestClient creates a new REST client
func NewRestClient(baseURL, token string) *RestClient {
	return &RestClient{
		BaseURL:   baseURL,
		Token:     token,
		Timeout:   DefaultTimeout,
		VerifySSL: true,
	}
}

// NewRestClientWithOptions creates a REST client with custom options
func NewRestClientWithOptions(baseURL, token string, timeout time.Duration, verifySSL bool) *RestClient {
	return &RestClient{
		BaseURL:   baseURL,
		Token:     token,
		Timeout:   timeout,
		VerifySSL: verifySSL,
	}
}

func (c *RestClient) getClient() *resty.Client {
	c.clientMu.Do(func() {
		c.client = resty.New()
		c.client.SetTimeout(c.Timeout)
		c.client.SetBaseURL(c.BaseURL)
		c.client.SetHeader("Authorization", "Bearer "+c.Token)
		c.client.SetHeader("Content-Type", "application/json")
		c.client.SetHeader("Accept", "application/json")

		if cfg := tlsConfig(c.VerifySSL); cfg != nil {
			c.client.SetTLSClientConfig(cfg)
		}

		// Response logging
		c.client.OnAfterResponse(func(client *resty.Client, resp *resty.Response) error {
			log.WithFields(log.Fields{
				"status":     resp.StatusCode(),
				"time":       resp.Time(),
				"url":        resp.Request.URL,
				"method":     resp.Request.Method,
				"bodyLength": len(resp.Body()),
			}).Debug("REST response")
			return nil
		})
	})
	return c.client
}

// Get makes a GET request
func (c *RestClient) Get(endpoint string) (interface{}, error) {
	return c.request("GET", endpoint, nil)
}

// Post makes a POST request
func (c *RestClient) Post(endpoint string, body interface{}) (interface{}, error) {
	return c.request("POST", endpoint, body)
}

// Put makes a PUT request
func (c *RestClient) Put(endpoint string, body interface{}) (interface{}, error) {
	return c.request("PUT", endpoint, body)
}

// Delete makes a DELETE request
func (c *RestClient) Delete(endpoint string) (interface{}, error) {
	return c.request("DELETE", endpoint, nil)
}

func (c *RestClient) request(method, endpoint string, body interface{}) (interface{}, error) {
	url := fmt.Sprintf("/api/%s", endpoint)

	req := c.getClient().R()

	if body != nil {
		req.SetBody(body)
	}

	log.WithFields(log.Fields{
		"method": method,
		"url":    c.BaseURL + url,
	}).Debug("REST request")

	var resp *resty.Response
	var err error

	switch method {
	case "GET":
		resp, err = req.Get(url)
	case "POST":
		resp, err = req.Post(url)
	case "PUT":
		resp, err = req.Put(url)
	case "DELETE":
		resp, err = req.Delete(url)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return c.handleResponse(resp)
}

func (c *RestClient) handleResponse(resp *resty.Response) (interface{}, error) {
	if resp.StatusCode() == http.StatusNoContent {
		return nil, nil
	}

	// Check for errors
	if resp.StatusCode() >= 400 {
		return nil, c.handleError(resp)
	}

	// Check content type
	contentType := resp.Header().Get("Content-Type")
	if contentType == "" || !isJSONContentType(contentType) {
		// Return as string
		return string(resp.Body()), nil
	}

	// Parse JSON
	var result interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

func (c *RestClient) handleError(resp *resty.Response) error {
	statusCode := resp.StatusCode()

	// Try to extract structured error message from JSON first,
	// falling back to the raw response body.
	var errResp struct {
		Message string `json:"message"`
	}
	var body string
	if err := json.Unmarshal(resp.Body(), &errResp); err == nil && errResp.Message != "" {
		body = errResp.Message
	} else {
		body = string(resp.Body())
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return NewError(ErrCodeAuthenticationError, "Authentication failed: "+body)
	case http.StatusForbidden:
		return NewError(ErrCodePermissionDenied, "Permission denied: "+body)
	case http.StatusNotFound:
		return NewError(ErrCodeNotFound, "Resource not found: "+body)
	case http.StatusBadRequest:
		return NewError(ErrCodeValidationError, "Bad request: "+body)
	default:
		return NewError(ErrCodeAPIError, fmt.Sprintf("API error (%d): %s", statusCode, body))
	}
}

func isJSONContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "application/json")
}

// ---------------------------------------------------------------------------
// Typed response helpers – eliminate the repeated "call + type-assert +
// unexpected response type" boilerplate in every high-level method.
// ---------------------------------------------------------------------------

// getMap issues a GET and asserts the result is a map.
func (c *RestClient) getMap(endpoint string) (map[string]interface{}, error) {
	result, err := c.Get(endpoint)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errUnexpectedResponse
}

// getList issues a GET and asserts the result is a slice.
func (c *RestClient) getList(endpoint string) ([]interface{}, error) {
	result, err := c.Get(endpoint)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, errUnexpectedResponse
}

// postMap issues a POST and asserts the result is a map.
func (c *RestClient) postMap(endpoint string, body interface{}) (map[string]interface{}, error) {
	result, err := c.Post(endpoint, body)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errUnexpectedResponse
}

// High-level API methods

// GetConfig returns the Home Assistant configuration
func (c *RestClient) GetConfig() (map[string]interface{}, error) {
	return c.getMap("config")
}

// GetStates returns all entity states
func (c *RestClient) GetStates() ([]interface{}, error) {
	return c.getList("states")
}

// GetState returns the state of a specific entity
func (c *RestClient) GetState(entityID string) (map[string]interface{}, error) {
	return c.getMap("states/" + entityID)
}

// GetServices returns all available services
func (c *RestClient) GetServices() ([]interface{}, error) {
	return c.getList("services")
}

// CallService calls a service
func (c *RestClient) CallService(domain, service string, data map[string]interface{}) (interface{}, error) {
	endpoint := fmt.Sprintf("services/%s/%s", domain, service)
	return c.Post(endpoint, data)
}

// CheckConfig validates the Home Assistant configuration
func (c *RestClient) CheckConfig() (map[string]interface{}, error) {
	return c.postMap("config/core/check_config", nil)
}

// Restart restarts Home Assistant
func (c *RestClient) Restart() error {
	_, err := c.Post("services/homeassistant/restart", nil)
	return err
}

// GetErrorLog returns the HA system error log.
// HA 2025.10+ (with supervisor) uses /api/hassio/core/logs/latest.
// Older installs use the now-removed /api/error_log endpoint.
// We try the new supervisor endpoint first and fall back to the legacy one.
func (c *RestClient) GetErrorLog() (string, error) {
	// Try supervisor endpoint (HA 2025.10+)
	result, err := c.Get("hassio/core/logs/latest")
	if err == nil {
		if s, ok := result.(string); ok {
			return s, nil
		}
		return fmt.Sprintf("%v", result), nil
	}

	// Fall back to legacy endpoint (pre-2024.4)
	result, err = c.Get("error_log")
	if err != nil {
		return "", err
	}
	if s, ok := result.(string); ok {
		return s, nil
	}
	return fmt.Sprintf("%v", result), nil
}

// GetHistory returns the state history for an entity
func (c *RestClient) GetHistory(entityID string, startTime, endTime string) ([]interface{}, error) {
	endpoint := "history/period"
	if startTime != "" {
		endpoint = "history/period/" + startTime
	}

	// Build query params
	q := url.Values{}
	if entityID != "" {
		q.Set("filter_entity_id", entityID)
	}
	if endTime != "" {
		q.Set("end_time", endTime)
	}
	if encoded := q.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	return c.getList(endpoint)
}

// Config Flow methods

// ConfigFlowCreate starts a new config flow for an integration
func (c *RestClient) ConfigFlowCreate(handler string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"handler":               handler,
		"show_advanced_options": false,
	}
	return c.postMap("config/config_entries/flow", body)
}

// ConfigFlowStep submits data to a config flow step
func (c *RestClient) ConfigFlowStep(flowID string, data map[string]interface{}) (map[string]interface{}, error) {
	return c.postMap("config/config_entries/flow/"+flowID, data)
}

// ConfigEntryDelete deletes a config entry by ID
func (c *RestClient) ConfigEntryDelete(entryID string) error {
	_, err := c.Delete("config/config_entries/entry/" + entryID)
	return err
}

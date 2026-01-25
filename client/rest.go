package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const (
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 * time.Second
)

// RestClient is an HTTP client for the Home Assistant REST API
type RestClient struct {
	BaseURL   string
	Token     string
	Timeout   time.Duration
	VerifySSL bool
	client    *resty.Client
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
	if c.client == nil {
		c.client = resty.New()
		c.client.SetTimeout(c.Timeout)
		c.client.SetBaseURL(c.BaseURL)
		c.client.SetHeader("Authorization", "Bearer "+c.Token)
		c.client.SetHeader("Content-Type", "application/json")
		c.client.SetHeader("Accept", "application/json")

		if !c.VerifySSL {
			c.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
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
	}
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
	body := string(resp.Body())

	// Try to parse error message from JSON
	var errResp struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(resp.Body(), &errResp); err == nil && errResp.Message != "" {
		body = errResp.Message
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return &APIError{Code: "AUTHENTICATION_ERROR", Message: "Authentication failed: " + body}
	case http.StatusForbidden:
		return &APIError{Code: "PERMISSION_DENIED", Message: "Permission denied: " + body}
	case http.StatusNotFound:
		return &APIError{Code: "NOT_FOUND", Message: "Resource not found: " + body}
	case http.StatusBadRequest:
		return &APIError{Code: "VALIDATION_ERROR", Message: "Bad request: " + body}
	default:
		return &APIError{Code: "API_ERROR", Message: fmt.Sprintf("API error (%d): %s", statusCode, body)}
	}
}

func isJSONContentType(contentType string) bool {
	return contentType == "application/json" ||
		len(contentType) > 16 && contentType[:16] == "application/json"
}

// APIError represents an API error
type APIError struct {
	Code    string
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

// High-level API methods

// GetConfig returns the Home Assistant configuration
func (c *RestClient) GetConfig() (map[string]interface{}, error) {
	result, err := c.Get("config")
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetStates returns all entity states
func (c *RestClient) GetStates() ([]interface{}, error) {
	result, err := c.Get("states")
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetState returns the state of a specific entity
func (c *RestClient) GetState(entityID string) (map[string]interface{}, error) {
	result, err := c.Get("states/" + entityID)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetServices returns all available services
func (c *RestClient) GetServices() ([]interface{}, error) {
	result, err := c.Get("services")
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// CallService calls a service
func (c *RestClient) CallService(domain, service string, data map[string]interface{}) (interface{}, error) {
	endpoint := fmt.Sprintf("services/%s/%s", domain, service)
	return c.Post(endpoint, data)
}

// CheckConfig validates the Home Assistant configuration
func (c *RestClient) CheckConfig() (map[string]interface{}, error) {
	result, err := c.Post("config/core/check_config", nil)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Restart restarts Home Assistant
func (c *RestClient) Restart() error {
	_, err := c.Post("services/homeassistant/restart", nil)
	return err
}

// GetErrorLog returns the error log
func (c *RestClient) GetErrorLog() (string, error) {
	result, err := c.Get("error_log")
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
	params := ""
	if entityID != "" {
		params = "?filter_entity_id=" + entityID
	}
	if endTime != "" {
		if params == "" {
			params = "?"
		} else {
			params += "&"
		}
		params += "end_time=" + endTime
	}

	result, err := c.Get(endpoint + params)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// Config Flow methods

// ConfigFlowCreate starts a new config flow for an integration
func (c *RestClient) ConfigFlowCreate(handler string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"handler":              handler,
		"show_advanced_options": false,
	}
	result, err := c.Post("config/config_entries/flow", body)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ConfigFlowStep submits data to a config flow step
func (c *RestClient) ConfigFlowStep(flowID string, data map[string]interface{}) (map[string]interface{}, error) {
	result, err := c.Post("config/config_entries/flow/"+flowID, data)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ConfigEntryDelete deletes a config entry by ID
func (c *RestClient) ConfigEntryDelete(entryID string) error {
	_, err := c.Delete("config/config_entries/entry/" + entryID)
	return err
}

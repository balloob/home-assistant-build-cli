package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WebSocketClient is a WebSocket client for the Home Assistant API
type WebSocketClient struct {
	URL           string
	Token         string
	Timeout       time.Duration
	VerifySSL     bool
	conn          *websocket.Conn
	messageID     int
	messageIDMu   sync.Mutex
	pending       map[int]chan *WSMessage
	pendingMu     sync.RWMutex
	subscriptions map[int]func(map[string]interface{})
	subsMu        sync.RWMutex
	done          chan struct{}
	authenticated bool
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	ID      int                    `json:"id,omitempty"`
	Type    string                 `json:"type"`
	Success bool                   `json:"success,omitempty"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *WSError               `json:"error,omitempty"`
	Event   map[string]interface{} `json:"event,omitempty"`

	// Fields for sending commands
	AccessToken string `json:"access_token,omitempty"`

	// Dynamic fields
	Extra map[string]interface{} `json:"-"`
}

// WSError represents a WebSocket error
type WSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(baseURL, token string) *WebSocketClient {
	wsURL, _ := BuildWebSocketURL(baseURL)
	return &WebSocketClient{
		URL:           wsURL,
		Token:         token,
		Timeout:       30 * time.Second,
		VerifySSL:     true,
		pending:       make(map[int]chan *WSMessage),
		subscriptions: make(map[int]func(map[string]interface{})),
	}
}

// Connect establishes the WebSocket connection and authenticates
func (c *WebSocketClient) Connect() error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	if !c.VerifySSL {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	log.WithField("url", c.URL).Debug("Connecting to WebSocket")

	conn, resp, err := dialer.Dial(c.URL, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("websocket connection failed (%d): %w", resp.StatusCode, err)
		}
		return fmt.Errorf("websocket connection failed: %w", err)
	}
	c.conn = conn

	// Read auth_required message
	msg, err := c.readMessage()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to read auth_required: %w", err)
	}
	if msg.Type != "auth_required" {
		c.conn.Close()
		return fmt.Errorf("unexpected message type: %s", msg.Type)
	}

	log.Debug("Received auth_required, sending auth")

	// Send authentication
	authMsg := map[string]string{
		"type":         "auth",
		"access_token": c.Token,
	}
	if err := c.conn.WriteJSON(authMsg); err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to send auth: %w", err)
	}

	// Read auth result
	msg, err = c.readMessage()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to read auth result: %w", err)
	}

	if msg.Type == "auth_invalid" {
		c.conn.Close()
		errMsg := "authentication failed"
		if msg.Error != nil {
			errMsg = msg.Error.Message
		}
		return fmt.Errorf("%s", errMsg)
	}
	if msg.Type != "auth_ok" {
		c.conn.Close()
		return fmt.Errorf("unexpected auth response: %s", msg.Type)
	}

	log.Debug("WebSocket authenticated successfully")

	c.authenticated = true
	c.done = make(chan struct{})

	// Start receive loop
	go c.receiveLoop()

	return nil
}

// Close closes the WebSocket connection
func (c *WebSocketClient) Close() error {
	c.authenticated = false

	if c.done != nil {
		close(c.done)
	}

	// Cancel pending requests
	c.pendingMu.Lock()
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
	c.pendingMu.Unlock()

	c.subsMu.Lock()
	c.subscriptions = make(map[int]func(map[string]interface{}))
	c.subsMu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *WebSocketClient) nextID() int {
	c.messageIDMu.Lock()
	defer c.messageIDMu.Unlock()
	c.messageID++
	return c.messageID
}

func (c *WebSocketClient) readMessage() (*WSMessage, error) {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	// Also parse into a generic map for extra fields
	var extra map[string]interface{}
	if err := json.Unmarshal(data, &extra); err == nil {
		msg.Extra = extra
	}

	return &msg, nil
}

func (c *WebSocketClient) receiveLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
		}

		msg, err := c.readMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return
			}
			log.WithError(err).Debug("WebSocket read error")
			continue
		}

		c.handleMessage(msg)
	}
}

func (c *WebSocketClient) handleMessage(msg *WSMessage) {
	switch msg.Type {
	case "result":
		c.pendingMu.RLock()
		ch, ok := c.pending[msg.ID]
		c.pendingMu.RUnlock()

		if ok {
			ch <- msg
			c.pendingMu.Lock()
			delete(c.pending, msg.ID)
			c.pendingMu.Unlock()
		}

	case "event":
		c.subsMu.RLock()
		callback, ok := c.subscriptions[msg.ID]
		c.subsMu.RUnlock()

		if ok && msg.Event != nil {
			callback(msg.Event)
		}

	case "pong":
		c.pendingMu.RLock()
		ch, ok := c.pending[msg.ID]
		c.pendingMu.RUnlock()

		if ok {
			ch <- msg
			c.pendingMu.Lock()
			delete(c.pending, msg.ID)
			c.pendingMu.Unlock()
		}
	}
}

// SendCommand sends a command and waits for a response
func (c *WebSocketClient) SendCommand(cmdType string, params map[string]interface{}) (interface{}, error) {
	if !c.authenticated {
		return nil, fmt.Errorf("not connected")
	}

	msgID := c.nextID()

	// Build message
	msg := map[string]interface{}{
		"id":   msgID,
		"type": cmdType,
	}
	for k, v := range params {
		msg[k] = v
	}

	// Create response channel
	respCh := make(chan *WSMessage, 1)
	c.pendingMu.Lock()
	c.pending[msgID] = respCh
	c.pendingMu.Unlock()

	log.WithFields(log.Fields{
		"id":   msgID,
		"type": cmdType,
	}).Debug("Sending WebSocket command")

	// Send message
	if err := c.conn.WriteJSON(msg); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respCh:
		if resp == nil {
			return nil, fmt.Errorf("connection closed")
		}
		if !resp.Success {
			errMsg := "unknown error"
			if resp.Error != nil {
				errMsg = resp.Error.Message
			}
			return nil, fmt.Errorf("%s", errMsg)
		}
		return resp.Result, nil

	case <-time.After(c.Timeout):
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("command timed out")
	}
}

// High-level API methods

// GetStates returns all entity states
func (c *WebSocketClient) GetStates() ([]interface{}, error) {
	result, err := c.SendCommand("get_states", nil)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetConfig returns the Home Assistant configuration
func (c *WebSocketClient) GetConfig() (map[string]interface{}, error) {
	result, err := c.SendCommand("get_config", nil)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetServices returns all available services
func (c *WebSocketClient) GetServices() (map[string]interface{}, error) {
	result, err := c.SendCommand("get_services", nil)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// CallService calls a service
func (c *WebSocketClient) CallService(domain, service string, data, target map[string]interface{}, returnResponse bool) (interface{}, error) {
	params := map[string]interface{}{
		"domain":  domain,
		"service": service,
	}
	if data != nil {
		params["service_data"] = data
	}
	if target != nil {
		params["target"] = target
	}
	if returnResponse {
		params["return_response"] = true
	}
	return c.SendCommand("call_service", params)
}

// Ping sends a ping message
func (c *WebSocketClient) Ping() error {
	_, err := c.SendCommand("ping", nil)
	return err
}

// Registry operations

// AreaRegistryList returns all areas
func (c *WebSocketClient) AreaRegistryList() ([]interface{}, error) {
	result, err := c.SendCommand("config/area_registry/list", nil)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// AreaRegistryCreate creates a new area
func (c *WebSocketClient) AreaRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{"name": name}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("config/area_registry/create", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// AreaRegistryUpdate updates an area
func (c *WebSocketClient) AreaRegistryUpdate(areaID string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{"area_id": areaID}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("config/area_registry/update", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// AreaRegistryDelete deletes an area
func (c *WebSocketClient) AreaRegistryDelete(areaID string) error {
	_, err := c.SendCommand("config/area_registry/delete", map[string]interface{}{
		"area_id": areaID,
	})
	return err
}

// FloorRegistryList returns all floors
func (c *WebSocketClient) FloorRegistryList() ([]interface{}, error) {
	result, err := c.SendCommand("config/floor_registry/list", nil)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// FloorRegistryCreate creates a new floor
func (c *WebSocketClient) FloorRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{"name": name}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("config/floor_registry/create", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// FloorRegistryUpdate updates a floor
func (c *WebSocketClient) FloorRegistryUpdate(floorID string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{"floor_id": floorID}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("config/floor_registry/update", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// FloorRegistryDelete deletes a floor
func (c *WebSocketClient) FloorRegistryDelete(floorID string) error {
	_, err := c.SendCommand("config/floor_registry/delete", map[string]interface{}{
		"floor_id": floorID,
	})
	return err
}

// LabelRegistryList returns all labels
func (c *WebSocketClient) LabelRegistryList() ([]interface{}, error) {
	result, err := c.SendCommand("config/label_registry/list", nil)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// LabelRegistryCreate creates a new label
func (c *WebSocketClient) LabelRegistryCreate(name string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{"name": name}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("config/label_registry/create", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// LabelRegistryUpdate updates a label
func (c *WebSocketClient) LabelRegistryUpdate(labelID string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{"label_id": labelID}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("config/label_registry/update", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// LabelRegistryDelete deletes a label
func (c *WebSocketClient) LabelRegistryDelete(labelID string) error {
	_, err := c.SendCommand("config/label_registry/delete", map[string]interface{}{
		"label_id": labelID,
	})
	return err
}

// DeviceRegistryList returns all devices
func (c *WebSocketClient) DeviceRegistryList() ([]interface{}, error) {
	result, err := c.SendCommand("config/device_registry/list", nil)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeviceRegistryUpdate updates a device
func (c *WebSocketClient) DeviceRegistryUpdate(deviceID string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{"device_id": deviceID}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("config/device_registry/update", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// EntityRegistryList returns all entities
func (c *WebSocketClient) EntityRegistryList() ([]interface{}, error) {
	result, err := c.SendCommand("config/entity_registry/list", nil)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// EntityRegistryGet returns a specific entity
func (c *WebSocketClient) EntityRegistryGet(entityID string) (map[string]interface{}, error) {
	result, err := c.SendCommand("config/entity_registry/get", map[string]interface{}{
		"entity_id": entityID,
	})
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// EntityRegistryUpdate updates an entity
func (c *WebSocketClient) EntityRegistryUpdate(entityID string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{"entity_id": entityID}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("config/entity_registry/update", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ZoneList returns all zones
func (c *WebSocketClient) ZoneList() ([]interface{}, error) {
	result, err := c.SendCommand("zone/list", nil)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ZoneCreate creates a new zone
func (c *WebSocketClient) ZoneCreate(name string, latitude, longitude, radius float64, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{
		"name":      name,
		"latitude":  latitude,
		"longitude": longitude,
		"radius":    radius,
	}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("zone/create", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ZoneUpdate updates a zone
func (c *WebSocketClient) ZoneUpdate(zoneID string, params map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{"zone_id": zoneID}
	for k, v := range params {
		p[k] = v
	}
	result, err := c.SendCommand("zone/update", p)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ZoneDelete deletes a zone
func (c *WebSocketClient) ZoneDelete(zoneID string) error {
	_, err := c.SendCommand("zone/delete", map[string]interface{}{
		"zone_id": zoneID,
	})
	return err
}

// SystemHealthInfo returns system health information using subscription
func (c *WebSocketClient) SystemHealthInfo() (map[string]interface{}, error) {
	if !c.authenticated {
		return nil, fmt.Errorf("not connected")
	}

	msgID := c.nextID()

	// Channel to receive events
	eventCh := make(chan map[string]interface{}, 100)
	doneCh := make(chan struct{})

	// Accumulated data
	data := make(map[string]interface{})
	var dataErr error

	// Register subscription callback
	c.subsMu.Lock()
	c.subscriptions[msgID] = func(event map[string]interface{}) {
		select {
		case eventCh <- event:
		default:
		}
	}
	c.subsMu.Unlock()

	// Cleanup function
	defer func() {
		c.subsMu.Lock()
		delete(c.subscriptions, msgID)
		c.subsMu.Unlock()
	}()

	// Send subscribe message
	msg := map[string]interface{}{
		"id":   msgID,
		"type": "system_health/info",
	}

	// Create response channel for the initial result
	respCh := make(chan *WSMessage, 1)
	c.pendingMu.Lock()
	c.pending[msgID] = respCh
	c.pendingMu.Unlock()

	if err := c.conn.WriteJSON(msg); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Wait for initial result (subscription confirmation)
	select {
	case resp := <-respCh:
		if resp == nil {
			return nil, fmt.Errorf("connection closed")
		}
		if !resp.Success {
			errMsg := "unknown error"
			if resp.Error != nil {
				errMsg = resp.Error.Message
			}
			return nil, fmt.Errorf("%s", errMsg)
		}
	case <-time.After(c.Timeout):
		return nil, fmt.Errorf("timeout waiting for subscription confirmation")
	}

	// Process events in a goroutine
	go func() {
		defer close(doneCh)
		for event := range eventCh {
			eventType, _ := event["type"].(string)

			switch eventType {
			case "initial":
				if eventData, ok := event["data"].(map[string]interface{}); ok {
					for k, v := range eventData {
						data[k] = v
					}
				}
			case "update":
				domain, _ := event["domain"].(string)
				key, _ := event["key"].(string)
				success, _ := event["success"].(bool)

				if domain != "" && key != "" {
					if _, exists := data[domain]; !exists {
						data[domain] = map[string]interface{}{
							"info": make(map[string]interface{}),
						}
					}
					if domainData, ok := data[domain].(map[string]interface{}); ok {
						if _, exists := domainData["info"]; !exists {
							domainData["info"] = make(map[string]interface{})
						}
						if infoData, ok := domainData["info"].(map[string]interface{}); ok {
							if success {
								infoData[key] = event["data"]
							} else {
								if errData, ok := event["error"].(map[string]interface{}); ok {
									infoData[key] = map[string]interface{}{
										"error": true,
										"value": errData["msg"],
									}
								}
							}
						}
					}
				}
			case "finish":
				return
			}
		}
	}()

	// Wait for finish or timeout
	select {
	case <-doneCh:
		// Normal completion
	case <-time.After(30 * time.Second):
		dataErr = fmt.Errorf("timeout waiting for system health data")
	}

	close(eventCh)

	if dataErr != nil {
		return nil, dataErr
	}

	return data, nil
}

// SearchRelated returns related items for a given item type and ID
// itemType can be: area, automation, automation_blueprint, config_entry, device, entity, floor, group, label, scene, script, script_blueprint
func (c *WebSocketClient) SearchRelated(itemType, itemID string) (map[string][]string, error) {
	result, err := c.SendCommand("search/related", map[string]interface{}{
		"item_type": itemType,
		"item_id":   itemID,
	})
	if err != nil {
		return nil, err
	}

	// Convert the result to a map of string slices
	resultMap := make(map[string][]string)
	if m, ok := result.(map[string]interface{}); ok {
		for key, value := range m {
			if arr, ok := value.([]interface{}); ok {
				var items []string
				for _, item := range arr {
					if str, ok := item.(string); ok {
						items = append(items, str)
					}
				}
				resultMap[key] = items
			}
		}
	}

	return resultMap, nil
}

// GetWebSocketClientForAuth creates a WebSocket client from auth manager
func GetWebSocketClientForAuth(baseURL, token string) *WebSocketClient {
	return NewWebSocketClient(baseURL, token)
}

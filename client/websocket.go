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
	writeMu       sync.Mutex // serialises conn.WriteJSON; gorilla/websocket does not allow concurrent writes
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
			return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("websocket connection failed (%d): %s", resp.StatusCode, err)}
		}
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("websocket connection failed: %s", err)}
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
		return &APIError{Code: ErrCodeAuthenticationError, Message: errMsg}
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
		return
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

	// Build message (without ID — assigned inside the write lock below)
	msg := map[string]interface{}{
		"type": cmdType,
	}
	for k, v := range params {
		msg[k] = v
	}

	// Assign ID and send atomically under the write lock.  Home Assistant
	// requires message IDs to arrive in strictly increasing order, so the
	// ID assignment and the write must happen in the same critical section
	// to prevent out-of-order delivery when multiple goroutines call
	// SendCommand concurrently.
	c.writeMu.Lock()
	msgID := c.nextID()
	msg["id"] = msgID

	respCh := make(chan *WSMessage, 1)
	c.pendingMu.Lock()
	c.pending[msgID] = respCh
	c.pendingMu.Unlock()

	log.WithFields(log.Fields{
		"id":   msgID,
		"type": cmdType,
	}).Debug("Sending WebSocket command")

	writeErr := c.conn.WriteJSON(msg)
	c.writeMu.Unlock()
	if writeErr != nil {
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("failed to send command: %w", writeErr)
	}

	// Wait for response
	select {
	case resp := <-respCh:
		if resp == nil {
			return nil, &APIError{Code: ErrCodeConnectionError, Message: "connection closed"}
		}
		if !resp.Success {
			code := ErrCodeAPIError
			errMsg := "unknown error"
			if resp.Error != nil {
				if resp.Error.Code != "" {
					code = resp.Error.Code
				}
				errMsg = resp.Error.Message
			}
			return nil, &APIError{Code: code, Message: errMsg}
		}
		return resp.Result, nil

	case <-time.After(c.Timeout):
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, &APIError{Code: ErrCodeTimeout, Message: "command timed out"}
	}
}

// sendListCommand sends a command and asserts the result is a []interface{}.
func (c *WebSocketClient) sendListCommand(cmdType string, params map[string]interface{}) ([]interface{}, error) {
	result, err := c.SendCommand(cmdType, params)
	if err != nil {
		return nil, err
	}
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}
	return nil, errUnexpectedResponse
}

// sendMapCommand sends a command and asserts the result is a map[string]interface{}.
func (c *WebSocketClient) sendMapCommand(cmdType string, params map[string]interface{}) (map[string]interface{}, error) {
	result, err := c.SendCommand(cmdType, params)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errUnexpectedResponse
}

// mergeAndSend builds a params map from a base key/value pair, merges in extra
// params, and sends the command expecting a map result.  This is the common
// pattern used by registry create/update methods.
func (c *WebSocketClient) mergeAndSend(cmdType, baseKey string, baseVal interface{}, extra map[string]interface{}) (map[string]interface{}, error) {
	p := map[string]interface{}{baseKey: baseVal}
	for k, v := range extra {
		p[k] = v
	}
	return c.sendMapCommand(cmdType, p)
}

// sendDelete sends a delete command with a single ID field and discards the
// result.  This is the common pattern used by registry delete methods.
func (c *WebSocketClient) sendDelete(cmdType, idField string, idVal interface{}) error {
	_, err := c.SendCommand(cmdType, map[string]interface{}{idField: idVal})
	return err
}



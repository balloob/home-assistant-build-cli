package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
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
	messageID     atomic.Int64
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
	return c.ConnectContext(context.Background())
}

// ConnectContext bounds dialing and authentication, including a silent peer.
func (c *WebSocketClient) ConnectContext(ctx context.Context) (err error) {
	defer func() {
		if err != nil && ctx.Err() != nil {
			err = ctx.Err()
		}
	}()
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	dialer.TLSClientConfig = tlsConfig(c.VerifySSL)

	log.WithField("url", c.URL).Debug("Connecting to WebSocket")

	conn, resp, err := dialer.DialContext(ctx, c.URL, nil)
	if err != nil {
		if resp != nil {
			return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("websocket connection failed (%d): %s", resp.StatusCode, err), Category: "connection", Retryable: true, Transport: "websocket", StatusCode: resp.StatusCode, SuggestedFix: "Verify URL reachability and Home Assistant network settings, then retry."}
		}
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("websocket connection failed: %s", err), Category: "connection", Retryable: true, Transport: "websocket", SuggestedFix: "Verify URL reachability and Home Assistant network settings, then retry."}
	}
	c.conn = conn
	defer func() {
		if err != nil {
			// Preserve the connection/authentication failure if cleanup also fails.
			_ = conn.Close()
		}
	}()
	deadline := time.Now().Add(c.Timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetReadDeadline(deadline); err != nil {
		return err
	}
	if err := conn.SetWriteDeadline(deadline); err != nil {
		return err
	}
	stop := context.AfterFunc(ctx, func() {
		// Interrupt authentication reads when the caller cancels.
		// The caller receives the context error, not a cleanup error.
		_ = conn.Close()
	})
	defer stop()

	// Read auth_required message
	msg, err := c.readMessage()
	if err != nil {
		c.conn.Close()
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to read auth_required: %s", err), Category: "connection", Retryable: true, Transport: "websocket"}
	}
	if msg.Type != "auth_required" {
		c.conn.Close()
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("unexpected message type: %s", msg.Type), Category: "protocol", Transport: "websocket"}
	}

	log.Debug("Received auth_required, sending auth")

	// Send authentication
	authMsg := map[string]string{
		"type":         "auth",
		"access_token": c.Token,
	}
	if err := c.conn.WriteJSON(authMsg); err != nil {
		c.conn.Close()
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to send auth: %s", err), Category: "connection", Retryable: true, Transport: "websocket"}
	}

	// Read auth result
	msg, err = c.readMessage()
	if err != nil {
		c.conn.Close()
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to read auth result: %s", err), Category: "connection", Retryable: true, Transport: "websocket"}
	}

	if msg.Type == "auth_invalid" {
		c.conn.Close()
		errMsg := "authentication failed"
		if msg.Error != nil {
			errMsg = msg.Error.Message
		}
		return &APIError{Code: ErrCodeAuthenticationError, Message: errMsg, Category: "authentication", Retryable: false, Transport: "websocket", SuggestedFix: "Run auth status and login again before retrying WebSocket commands."}
	}
	if msg.Type != "auth_ok" {
		c.conn.Close()
		return &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("unexpected auth response: %s", msg.Type), Category: "protocol", Transport: "websocket"}
	}

	log.Debug("WebSocket authenticated successfully")
	if err := ctx.Err(); err != nil {
		conn.Close()
		return err
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		conn.Close()
		return err
	}
	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		conn.Close()
		return err
	}

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
	return int(c.messageID.Add(1))
}

func (c *WebSocketClient) readMessage() (*WSMessage, error) {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	// Keep result numbers lossless until the caller chooses a decoding mode.
	var msg struct {
		WSMessage
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	msg.WSMessage.Result = msg.Result
	return &msg.WSMessage, nil
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
	case "result", "pong":
		c.pendingMu.Lock()
		ch, ok := c.pending[msg.ID]
		if ok {
			ch <- msg
			delete(c.pending, msg.ID)
		}
		c.pendingMu.Unlock()

	case "event":
		c.subsMu.RLock()
		callback, ok := c.subscriptions[msg.ID]
		c.subsMu.RUnlock()

		if ok && msg.Event != nil {
			callback(msg.Event)
		}

	}
}

// requireAuth returns an error if the client is not authenticated.
func (c *WebSocketClient) requireAuth() error {
	if !c.authenticated {
		return &APIError{Code: ErrCodeConnectionError, Message: "not connected"}
	}
	return nil
}

// SendCommand sends a command and waits for a response
func (c *WebSocketClient) SendCommand(cmdType string, params map[string]interface{}) (interface{}, error) {
	return c.sendCommandContext(context.Background(), cmdType, params, false)
}

// SendCommandContext sends once. Cancellation after the write does not imply
// that Home Assistant cancelled or rejected the operation.
func (c *WebSocketClient) SendCommandContext(ctx context.Context, cmdType string, params map[string]interface{}) (interface{}, error) {
	return c.sendCommandContext(ctx, cmdType, params, true)
}

func (c *WebSocketClient) sendCommandContext(ctx context.Context, cmdType string, params map[string]interface{}, exactNumbers bool) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := c.requireAuth(); err != nil {
		return nil, err
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
	if err := ctx.Err(); err != nil {
		c.writeMu.Unlock()
		return nil, err
	}
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

	deadline := time.Now().Add(c.Timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	writeErr := c.conn.SetWriteDeadline(deadline)
	if writeErr == nil {
		writeErr = c.conn.WriteJSON(msg)
	}
	// Other operations (including subscriptions) share this connection.
	if resetErr := c.conn.SetWriteDeadline(time.Time{}); writeErr == nil {
		writeErr = resetErr
	}
	c.writeMu.Unlock()
	if writeErr != nil {
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to send command: %s", writeErr), Category: "connection", Retryable: true, Transport: "websocket"}
	}

	// Wait for response
	timer := time.NewTimer(c.Timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, ctx.Err()
	case resp := <-respCh:
		if resp == nil {
			return nil, &APIError{Code: ErrCodeConnectionError, Message: "connection closed", Category: "connection", Retryable: true, Transport: "websocket", SuggestedFix: "Reconnect and retry the command."}
		}
		if !resp.Success {
			return nil, wsResponseError(resp)
		}
		raw, ok := resp.Result.(json.RawMessage)
		if !ok {
			return resp.Result, nil
		}
		if len(raw) == 0 {
			return nil, nil
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		if exactNumbers {
			decoder.UseNumber()
		}
		var result any
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}
		return result, nil

	case <-timer.C:
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, &APIError{Code: ErrCodeTimeout, Message: "command timed out", Category: "timeout", Retryable: true, Transport: "websocket", SuggestedFix: "Retry after a short delay or increase timeout for long-running operations."}
	}
}

// SubscribeEvents subscribes to Home Assistant events of eventType. Events
// arrive on the returned channel until the returned function is called.
func (c *WebSocketClient) SubscribeEvents(eventType string) (<-chan map[string]interface{}, func(), error) {
	if err := c.requireAuth(); err != nil {
		return nil, nil, err
	}

	eventCh := make(chan map[string]interface{}, 10)

	// Register the callback before the subscription is sent, so that no
	// event that arrives right after the result is lost.
	c.writeMu.Lock()
	msgID := c.nextID()

	c.subsMu.Lock()
	c.subscriptions[msgID] = func(event map[string]interface{}) {
		select {
		case eventCh <- event:
		default:
		}
	}
	c.subsMu.Unlock()

	respCh := make(chan *WSMessage, 1)
	c.pendingMu.Lock()
	c.pending[msgID] = respCh
	c.pendingMu.Unlock()

	writeErr := c.conn.WriteJSON(map[string]interface{}{
		"id":         msgID,
		"type":       "subscribe_events",
		"event_type": eventType,
	})
	c.writeMu.Unlock()

	removeSubscription := func() {
		c.subsMu.Lock()
		delete(c.subscriptions, msgID)
		c.subsMu.Unlock()
	}

	if writeErr != nil {
		removeSubscription()
		c.pendingMu.Lock()
		delete(c.pending, msgID)
		c.pendingMu.Unlock()
		return nil, nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to send command: %s", writeErr), Category: "connection", Retryable: true, Transport: "websocket"}
	}

	select {
	case resp := <-respCh:
		if resp == nil {
			removeSubscription()
			return nil, nil, &APIError{Code: ErrCodeConnectionError, Message: "connection closed", Category: "connection", Retryable: true, Transport: "websocket"}
		}
		if !resp.Success {
			removeSubscription()
			return nil, nil, wsResponseError(resp)
		}
	case <-time.After(c.Timeout):
		removeSubscription()
		return nil, nil, &APIError{Code: ErrCodeTimeout, Message: "timeout waiting for subscription confirmation", Category: "timeout", Retryable: true, Transport: "websocket"}
	}

	unsubscribe := func() {
		removeSubscription()
		_, _ = c.SendCommand("unsubscribe_events", map[string]interface{}{"subscription": msgID})
	}
	return eventCh, unsubscribe, nil
}

// wsResponseError converts a failed WSMessage into an *APIError.
// It extracts the code and message from the WSError field, falling back to
// defaults when the fields are absent.
func wsResponseError(resp *WSMessage) *APIError {
	code := ErrCodeAPIError
	errMsg := "unknown error"
	if resp.Error != nil {
		if resp.Error.Code != "" {
			code = resp.Error.Code
		}
		errMsg = resp.Error.Message
	}
	return &APIError{Code: code, Message: errMsg, Category: "api", Retryable: false, Transport: "websocket"}
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

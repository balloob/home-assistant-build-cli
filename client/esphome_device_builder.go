package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// The ESPHome Device Builder (the dashboard's successor) replaced most of the
// classic dashboard's REST endpoints with one multiplexed WebSocket at /ws.
// See https://github.com/esphome/device-builder/blob/main/docs/API.md.

// errNotDeviceBuilder means the server has no /ws API, so it is the classic
// dashboard and the caller should use the legacy endpoints.
var errNotDeviceBuilder = errors.New("not an ESPHome Device Builder")

// deviceBuilderStreamCommands maps classic dashboard streaming WebSocket paths
// to their Device Builder /ws commands. /compile and /upload are kept by the
// Device Builder as legacy endpoints, so they are not listed here.
var deviceBuilderStreamCommands = map[string]string{
	"/logs":     "devices/logs",
	"/validate": "devices/validate",
}

type deviceBuilderFrame struct {
	MessageID string          `json:"message_id"`
	Result    json.RawMessage `json:"result"`
	ErrorCode string          `json:"error_code"`
	Details   string          `json:"details"`
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
}

// deviceBuilderCall runs one command on the Device Builder /ws API and returns
// its result. For streaming commands, onEvent receives every event frame
// before the final result. It returns errNotDeviceBuilder when the server
// does not speak the /ws protocol, and caches that so later calls skip /ws.
func (c *ESPHomeClient) deviceBuilderCall(command string, args map[string]any, onEvent func(event string, data json.RawMessage)) (json.RawMessage, error) {
	if c.legacyDashboard {
		return nil, errNotDeviceBuilder
	}

	wsURL, err := c.buildWSURL("/ws")
	if err != nil {
		return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to build WebSocket URL: %s", err)}
	}

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second, TLSClientConfig: tlsConfig(c.VerifySSL)}
	headers := http.Header{}
	if c.Token != "" {
		headers.Set("Authorization", "Bearer "+c.Token)
	}
	if c.IngressSession != "" {
		headers.Set("Cookie", "ingress_session="+c.IngressSession)
	}

	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		// The classic dashboard answers /ws with 404; let the legacy endpoints
		// report any real connection problem.
		log.WithError(err).Debug("ESPHome /ws unavailable, using classic dashboard endpoints")
		c.legacyDashboard = true
		return nil, errNotDeviceBuilder
	}
	defer conn.Close()

	streaming := onEvent != nil
	// Streams wait only for the greeting; a zero Timeout must not expire at once.
	timeout := 10 * time.Second
	if !streaming && c.Timeout > 0 {
		timeout = c.Timeout
	}
	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	// The server greets every connection with a ServerInfoMessage.
	var info struct {
		ServerVersion *string `json:"server_version"`
		RequiresAuth  bool    `json:"requires_auth"`
	}
	if err := conn.ReadJSON(&info); err != nil || info.ServerVersion == nil {
		// A classic dashboard behind ingress closes at once; silence is a slow
		// Device Builder, which must not be sent to the legacy endpoints.
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("ESPHome Device Builder did not answer on /ws: %s", err)}
		}
		c.legacyDashboard = true
		return nil, errNotDeviceBuilder
	}
	if info.RequiresAuth {
		return nil, &APIError{
			Code:    ErrCodeAuthenticationError,
			Message: "ESPHome Device Builder requires a login; reach it through Home Assistant ingress (unset HAB_ESPHOME_URL) or pass a Device Builder session token in HAB_ESPHOME_TOKEN",
		}
	}
	if streaming {
		_ = conn.SetReadDeadline(time.Time{})
	}

	const messageID = "1"
	if args == nil {
		args = map[string]any{}
	}
	log.WithFields(log.Fields{"command": command, "url": wsURL}).Debug("ESPHome Device Builder command")
	if err := conn.WriteJSON(map[string]any{"command": command, "message_id": messageID, "args": args}); err != nil {
		return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("failed to send %s: %s", command, err)}
	}

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("ESPHome Device Builder connection closed during %s: %s", command, err)}
		}
		var frame deviceBuilderFrame
		if err := json.Unmarshal(raw, &frame); err != nil {
			log.WithField("raw", string(raw)).Debug("Unparseable ESPHome Device Builder message")
			continue
		}
		// Errors for unparseable commands carry an empty message_id.
		if frame.MessageID != messageID && !(frame.MessageID == "" && frame.ErrorCode != "") {
			continue
		}
		switch {
		case frame.ErrorCode != "":
			return nil, deviceBuilderError(command, frame)
		case frame.Event != "":
			if onEvent != nil {
				onEvent(frame.Event, frame.Data)
			}
		default:
			return frame.Result, nil
		}
	}
}

func deviceBuilderError(command string, frame deviceBuilderFrame) error {
	code := ErrCodeAPIError
	switch frame.ErrorCode {
	case "not_found":
		code = ErrCodeNotFound
	case "invalid_args", "precondition_failed":
		code = ErrCodeValidationError
	case "not_authenticated":
		code = ErrCodeAuthenticationError
	}
	message := frame.Details
	if message == "" {
		message = fmt.Sprintf("ESPHome Device Builder %s failed: %s", command, frame.ErrorCode)
	}
	return &APIError{Code: code, Message: message, Details: map[string]any{
		"command":    command,
		"error_code": frame.ErrorCode,
	}}
}

// deviceBuilderStream runs a streaming /ws command and adapts its frames to
// the classic dashboard's line/exit events.
func (c *ESPHomeClient) deviceBuilderStream(command string, args map[string]any, callback func(ESPHomeStreamEvent)) (int, error) {
	exitCode := -1
	_, err := c.deviceBuilderCall(command, args, func(event string, data json.RawMessage) {
		switch event {
		case "output":
			var line string
			if json.Unmarshal(data, &line) == nil {
				// The Device Builder strips line terminators; the classic one kept them.
				callback(ESPHomeStreamEvent{Event: "line", Data: line + "\n"})
			}
		case "result":
			var result struct {
				Code *int `json:"code"`
			}
			if json.Unmarshal(data, &result) == nil && result.Code != nil {
				exitCode = *result.Code
			}
			callback(ESPHomeStreamEvent{Event: "exit", Code: &exitCode})
		}
	})
	if err != nil {
		return -1, err
	}
	return exitCode, nil
}

// streamDeviceBuilderRun emulates the classic /run (compile + upload) with the
// /compile and /upload endpoints the Device Builder still serves. The compile
// step's exit event is withheld so callers see a single exit.
func (c *ESPHomeClient) streamDeviceBuilderRun(spawnMsg map[string]any, callback func(ESPHomeStreamEvent)) (int, error) {
	compileMsg := map[string]any{"type": "spawn", "configuration": spawnMsg["configuration"]}
	code, err := c.streamLegacy("/compile", compileMsg, func(event ESPHomeStreamEvent) {
		if event.Event != "exit" {
			callback(event)
		}
	})
	if err != nil {
		return code, err
	}
	if code != 0 {
		callback(ESPHomeStreamEvent{Event: "exit", Code: &code})
		return code, nil
	}
	uploadMsg := map[string]any{"type": "spawn", "configuration": spawnMsg["configuration"], "port": spawnMsg["port"]}
	return c.streamLegacy("/upload", uploadMsg, callback)
}

// Troubleshoot runs the Device Builder's connectivity probe (DNS, mDNS, ping)
// for a device. The classic dashboard has no equivalent.
func (c *ESPHomeClient) Troubleshoot(configuration string) (map[string]any, error) {
	raw, err := c.deviceBuilderCall("devices/troubleshoot", map[string]any{"configuration": configuration}, nil)
	if errors.Is(err, errNotDeviceBuilder) {
		return nil, &APIError{Code: ErrCodeAPIError, Message: "troubleshoot requires the ESPHome Device Builder; this server is the classic ESPHome dashboard"}
	}
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, &APIError{Code: ErrCodeAPIError, Message: fmt.Sprintf("failed to parse troubleshoot response: %s", err)}
	}
	return result, nil
}

// isHTMLBody reports whether a REST response is a web page, which the Device
// Builder's single-page app returns with 200 OK for any unknown GET path.
func isHTMLBody(contentType string, body []byte) bool {
	if strings.HasPrefix(strings.ToLower(contentType), "text/html") {
		return true
	}
	start := strings.ToLower(strings.TrimSpace(string(body[:min(len(body), 64)])))
	return strings.HasPrefix(start, "<!doctype html") || strings.HasPrefix(start, "<html")
}

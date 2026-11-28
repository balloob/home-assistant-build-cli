package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// The ESPHome Device Builder serves its API as one multiplexed WebSocket at /ws.
// See https://github.com/esphome/device-builder/blob/main/docs/API.md.

// errLegacyDashboard is the message for a server that does not speak the /ws API.
const errLegacyDashboard = "this ESPHome server is not the ESPHome Device Builder; the legacy ESPHome dashboard is no longer supported, upgrade the ESPHome app"

type deviceBuilderFrame struct {
	MessageID string          `json:"message_id"`
	Result    json.RawMessage `json:"result"`
	ErrorCode string          `json:"error_code"`
	Details   string          `json:"details"`
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
}

// call runs one command on the Device Builder /ws API and returns its result.
// For streaming commands, onEvent receives every event frame before the final
// result.
func (c *ESPHomeClient) call(command string, args map[string]any, onEvent func(event string, data json.RawMessage)) (json.RawMessage, error) {
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

	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
			details := map[string]any{"status_code": resp.StatusCode}
			switch {
			case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
				return nil, &APIError{Code: ErrCodeAuthenticationError, Message: fmt.Sprintf("ESPHome Device Builder refused the connection with status %d; check the Home Assistant login or HAB_ESPHOME_TOKEN", resp.StatusCode), Details: details}
			case resp.StatusCode >= http.StatusInternalServerError:
				// Home Assistant ingress answers 5xx while the app is stopped or starting.
				return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("ESPHome Device Builder is not available (status %d); check that the ESPHome app is running", resp.StatusCode), Details: details}
			}
			// Any other HTTP answer without the upgrade: some other server owns this URL.
			return nil, &APIError{Code: ErrCodeAPIError, Message: errLegacyDashboard, Details: details}
		}
		return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("ESPHome Device Builder connection failed: %s", err)}
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
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, &APIError{Code: ErrCodeConnectionError, Message: fmt.Sprintf("ESPHome Device Builder did not answer on /ws: %s", err)}
		}
		return nil, &APIError{Code: ErrCodeAPIError, Message: errLegacyDashboard}
	}
	c.checked = true
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

// callInto runs a command and decodes its result into dest.
func (c *ESPHomeClient) callInto(command string, args map[string]any, dest any) error {
	raw, err := c.call(command, args, nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return &APIError{Code: ErrCodeAPIError, Message: fmt.Sprintf("failed to parse %s response: %s", command, err)}
	}
	return nil
}

func deviceBuilderError(command string, frame deviceBuilderFrame) error {
	code := ErrCodeAPIError
	switch frame.ErrorCode {
	case "not_found":
		code = ErrCodeNotFound
	case "invalid_args", "precondition_failed", "already_exists":
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

// GetDevices returns the list of configured and importable ESPHome devices.
func (c *ESPHomeClient) GetDevices() (*ESPHomeDeviceList, error) {
	var result struct {
		Configured []struct {
			ESPHomeDevice
			RuntimeState struct {
				State           string `json:"state"`
				DeployedVersion string `json:"deployed_version"`
			} `json:"runtime_state"`
		} `json:"configured"`
		Importable []ESPHomeImportableDevice `json:"importable"`
	}
	if err := c.callInto("devices/list", nil, &result); err != nil {
		return nil, err
	}
	list := &ESPHomeDeviceList{Configured: make([]ESPHomeDevice, 0, len(result.Configured)), Importable: result.Importable}
	for _, d := range result.Configured {
		device := d.ESPHomeDevice
		device.State = d.RuntimeState.State
		device.DeployedVersion = d.RuntimeState.DeployedVersion
		list.Configured = append(list.Configured, device)
	}
	if list.Importable == nil {
		list.Importable = []ESPHomeImportableDevice{}
	}
	return list, nil
}

// GetBoards returns the board catalog for a platform.
func (c *ESPHomeClient) GetBoards(platform string) ([]ESPHomeBoard, error) {
	var result struct {
		Boards []ESPHomeBoard `json:"boards"`
	}
	// The catalog is a few thousand boards at most; fetch it in one page.
	if err := c.callInto("boards/get_boards", map[string]any{"platform": strings.ToLower(platform), "limit": 100000}, &result); err != nil {
		return nil, err
	}
	boards := result.Boards
	if boards == nil {
		boards = []ESPHomeBoard{}
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
		ESPHomeVersion string `json:"esphome_version"`
	}
	if err := c.callInto("config/version", nil, &result); err != nil {
		return "", err
	}
	return result.ESPHomeVersion, nil
}

// GetSerialPorts lists the serial ports attached to the ESPHome server.
func (c *ESPHomeClient) GetSerialPorts() ([]ESPHomeSerialPort, error) {
	var result []ESPHomeSerialPort
	if err := c.callInto("config/serial_ports", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateConfig creates a new ESPHome configuration. With FileContent the YAML
// is written as-is; with Board it is generated from the board's template;
// with neither, a minimal stub is written.
func (c *ESPHomeClient) CreateConfig(req ESPHomeCreateRequest) (*ESPHomeCreateResponse, error) {
	args := map[string]any{"name": req.Name}
	if req.Board != "" {
		args["board_id"] = req.Board
	}
	if req.SSID != "" {
		args["ssid"] = req.SSID
	}
	if req.PSK != "" {
		args["psk"] = req.PSK
	}
	if len(req.FileContent) > 0 {
		args["file_content"] = string(req.FileContent)
	}
	var result ESPHomeCreateResponse
	if err := c.callInto("devices/create", args, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ImportConfig imports a package-backed ESPHome device configuration.
func (c *ESPHomeClient) ImportConfig(req ESPHomeImportRequest) (*ESPHomeImportResponse, error) {
	args := map[string]any{
		"name":               req.Name,
		"project_name":       req.ProjectName,
		"package_import_url": req.PackageImportURL,
	}
	if req.FriendlyName != "" {
		args["friendly_name"] = req.FriendlyName
	}
	if req.Encryption {
		// The Device Builder mints an API key for any non-empty value.
		args["encryption"] = "true"
	}
	var result ESPHomeImportResponse
	if err := c.callInto("devices/import", args, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ReadConfig reads the YAML configuration for a device.
func (c *ESPHomeClient) ReadConfig(configuration string) (string, error) {
	var content string
	if err := c.callInto("devices/get_config", map[string]any{"configuration": configuration}, &content); err != nil {
		return "", err
	}
	return content, nil
}

// WriteConfig writes YAML configuration for a device.
func (c *ESPHomeClient) WriteConfig(configuration, content string) error {
	return c.writeConfig(configuration, content, nil)
}

// WriteConfigIfUnchanged writes YAML configuration for a device only while the
// file still holds expected (the content read before patching). It fails with
// a validation error, writing nothing, when the file changed in the meantime.
func (c *ESPHomeClient) WriteConfigIfUnchanged(configuration, content, expected string) error {
	return c.writeConfig(configuration, content, &expected)
}

func (c *ESPHomeClient) writeConfig(configuration, content string, expected *string) error {
	args := map[string]any{"configuration": configuration, "content": content}
	if expected != nil {
		args["expected"] = *expected
	}
	_, err := c.call("devices/update_config", args, nil)
	var apiErr *APIError
	if expected != nil && errors.As(err, &apiErr) && apiErr.Details["error_code"] == "precondition_failed" {
		apiErr.Message = fmt.Sprintf("configuration %q changed since it was read; nothing was written. %s", configuration, apiErr.Message)
	}
	return err
}

// Troubleshoot runs the Device Builder's connectivity probe (DNS, mDNS, ping)
// for a device.
func (c *ESPHomeClient) Troubleshoot(configuration string) (map[string]any, error) {
	var result map[string]any
	if err := c.callInto("devices/troubleshoot", map[string]any{"configuration": configuration}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// StreamCommand runs a streaming operation on a configuration and reports its
// output through callback as "line" events, then one "exit" event. operation
// is one of "build", "upload", "run" (build then upload), "logs" or
// "validate"; port applies to upload, run and logs. It returns the exit code.
func (c *ESPHomeClient) StreamCommand(operation, configuration, port string, callback func(ESPHomeStreamEvent)) (int, error) {
	switch operation {
	case "logs":
		return c.streamProcess("devices/logs", map[string]any{"configuration": configuration, "port": port}, callback)
	case "validate":
		return c.streamProcess("devices/validate", map[string]any{"configuration": configuration}, callback)
	case "build":
		return c.streamJob("firmware/compile", map[string]any{"configuration": configuration}, callback)
	case "upload":
		return c.streamJob("firmware/upload", map[string]any{"configuration": configuration, "port": port}, callback)
	case "run":
		// Withhold the build's exit event so callers see a single exit.
		code, err := c.streamJob("firmware/compile", map[string]any{"configuration": configuration}, func(event ESPHomeStreamEvent) {
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
		return c.streamJob("firmware/upload", map[string]any{"configuration": configuration, "port": port}, callback)
	default:
		return -1, fmt.Errorf("unknown ESPHome stream operation %q", operation)
	}
}

// streamProcess follows a per-connection subprocess stream (logs, validate).
func (c *ESPHomeClient) streamProcess(command string, args map[string]any, callback func(ESPHomeStreamEvent)) (int, error) {
	exitCode := -1
	_, err := c.call(command, args, func(event string, data json.RawMessage) {
		switch event {
		case "output":
			var line string
			if json.Unmarshal(data, &line) == nil {
				// These streams strip line terminators.
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

// streamJob queues a firmware job and follows its output until it ends.
func (c *ESPHomeClient) streamJob(command string, args map[string]any, callback func(ESPHomeStreamEvent)) (int, error) {
	var job struct {
		JobID string `json:"job_id"`
	}
	if err := c.callInto(command, args, &job); err != nil {
		return -1, err
	}

	exitCode := -1
	_, err := c.call("firmware/follow_job", map[string]any{"job_id": job.JobID}, func(event string, data json.RawMessage) {
		switch event {
		case "output":
			var line string
			if json.Unmarshal(data, &line) == nil {
				// Job output keeps its own \n or \r terminator.
				callback(ESPHomeStreamEvent{Event: "line", Data: line})
			}
		case "result":
			var result struct {
				Status   string  `json:"status"`
				ExitCode *int    `json:"exit_code"`
				Error    *string `json:"error"`
			}
			if json.Unmarshal(data, &result) != nil {
				return
			}
			switch {
			case result.ExitCode != nil:
				exitCode = *result.ExitCode
			case result.Status == "completed":
				exitCode = 0
			default: // failed or cancelled before the tool exited
				exitCode = 1
			}
			if result.Status != "completed" && result.Error != nil && *result.Error != "" {
				callback(ESPHomeStreamEvent{Event: "line", Data: *result.Error + "\n"})
			}
			callback(ESPHomeStreamEvent{Event: "exit", Code: &exitCode})
		}
	})
	if err != nil {
		return -1, err
	}
	return exitCode, nil
}

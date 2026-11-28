package client

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type wsCommand struct {
	Command   string         `json:"command"`
	MessageID string         `json:"message_id"`
	Args      map[string]any `json:"args"`
}

// newDeviceBuilderServer emulates the ESPHome Device Builder: a /ws endpoint
// that greets with a ServerInfoMessage and passes each command to handle,
// plus any extra REST routes. It records every command received.
func newDeviceBuilderServer(t *testing.T, handle func(cmd wsCommand, send func(any)), extra map[string]http.HandlerFunc) (*httptest.Server, func() []wsCommand) {
	t.Helper()
	var mu sync.Mutex
	var received []wsCommand
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.WriteJSON(map[string]any{"server_version": "1.0.0", "esphome_version": "2026.11.1", "port": 6052, "requires_auth": false})
		var cmd wsCommand
		if err := conn.ReadJSON(&cmd); err != nil {
			return
		}
		mu.Lock()
		received = append(received, cmd)
		mu.Unlock()
		handle(cmd, func(v any) { _ = conn.WriteJSON(v) })
	})
	for path, h := range extra {
		mux.HandleFunc(path, h)
	}
	// Like the Device Builder's web app, answer any other GET with a page.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, "<!doctype html><title>ESPHome Device Builder</title>")
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, func() []wsCommand {
		mu.Lock()
		defer mu.Unlock()
		return append([]wsCommand(nil), received...)
	}
}

func TestDeviceBuilderReadConfig(t *testing.T) {
	server, received := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": "esphome:\n  name: kitchen\n"})
	}, nil)

	content, err := NewESPHomeClient(server.URL, "").ReadConfig("kitchen.yaml")
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if content != "esphome:\n  name: kitchen\n" {
		t.Fatalf("content = %q", content)
	}
	cmds := received()
	if len(cmds) != 1 || cmds[0].Command != "devices/get_config" || cmds[0].Args["configuration"] != "kitchen.yaml" {
		t.Fatalf("commands = %+v", cmds)
	}
}

func TestDeviceBuilderZeroTimeout(t *testing.T) {
	server, _ := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": "a: 1\n"})
	}, nil)

	c := &ESPHomeClient{BaseURL: server.URL} // zero Timeout
	if content, err := c.ReadConfig("kitchen.yaml"); err != nil || content != "a: 1\n" {
		t.Fatalf("ReadConfig = %q, %v", content, err)
	}
}

func TestDeviceBuilderReadConfigNotFound(t *testing.T) {
	server, _ := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "error_code": "not_found", "details": "Configuration missing.yaml not found"})
	}, nil)

	_, err := NewESPHomeClient(server.URL, "").ReadConfig("missing.yaml")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != ErrCodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

func TestDeviceBuilderWriteConfigIfUnchanged(t *testing.T) {
	server, received := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": nil})
	}, nil)

	if err := NewESPHomeClient(server.URL, "").WriteConfigIfUnchanged("kitchen.yaml", "new: 1\n", "old: 1\n"); err != nil {
		t.Fatalf("WriteConfigIfUnchanged: %v", err)
	}
	cmds := received()
	if len(cmds) != 1 || cmds[0].Command != "devices/update_config" {
		t.Fatalf("commands = %+v", cmds)
	}
	args := cmds[0].Args
	if args["configuration"] != "kitchen.yaml" || args["content"] != "new: 1\n" || args["expected"] != "old: 1\n" {
		t.Fatalf("args = %+v", args)
	}
}

func TestDeviceBuilderWriteConfigWithoutExpected(t *testing.T) {
	server, received := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": nil})
	}, nil)

	if err := NewESPHomeClient(server.URL, "").WriteConfig("kitchen.yaml", "new: 1\n"); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	if _, ok := received()[0].Args["expected"]; ok {
		t.Fatal("WriteConfig must not send expected")
	}
}

func TestDeviceBuilderWriteConfigConflict(t *testing.T) {
	server, _ := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "error_code": "precondition_failed", "details": "kitchen.yaml changed since it was read:\n-old: 1\n+other: 2"})
	}, nil)

	err := NewESPHomeClient(server.URL, "").WriteConfigIfUnchanged("kitchen.yaml", "new: 1\n", "old: 1\n")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != ErrCodeValidationError {
		t.Fatalf("err = %v, want VALIDATION_ERROR", err)
	}
	if !strings.Contains(apiErr.Message, "nothing was written") || !strings.Contains(apiErr.Message, "+other: 2") {
		t.Fatalf("message = %q", apiErr.Message)
	}
}

func TestDeviceBuilderLogs(t *testing.T) {
	server, received := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "event": "output", "data": "[I][app:100]: Running"})
		send(map[string]any{"message_id": "other", "event": "output", "data": "not ours"})
		send(map[string]any{"message_id": cmd.MessageID, "event": "output", "data": "[D][sensor:094]: 21.5"})
		send(map[string]any{"message_id": cmd.MessageID, "event": "result", "data": map[string]any{"success": false, "code": 2}})
		send(map[string]any{"message_id": cmd.MessageID, "result": nil})
	}, nil)

	var events []ESPHomeStreamEvent
	code, err := NewESPHomeClient(server.URL, "").StreamCommand("logs", "kitchen.yaml", "OTA", func(e ESPHomeStreamEvent) {
		events = append(events, e)
	})
	if err != nil {
		t.Fatalf("StreamCommand: %v", err)
	}
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if len(events) != 3 || events[0].Data != "[I][app:100]: Running\n" || events[1].Data != "[D][sensor:094]: 21.5\n" || events[2].Event != "exit" || *events[2].Code != 2 {
		t.Fatalf("events = %+v", events)
	}
	cmd := received()[0]
	if cmd.Command != "devices/logs" || cmd.Args["configuration"] != "kitchen.yaml" || cmd.Args["port"] != "OTA" {
		t.Fatalf("command = %+v", cmd)
	}
}

func TestDeviceBuilderTroubleshoot(t *testing.T) {
	server, _ := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{"configuration": cmd.Args["configuration"], "dns_resolved": true}})
	}, nil)

	result, err := NewESPHomeClient(server.URL, "").Troubleshoot("kitchen.yaml")
	if err != nil {
		t.Fatalf("Troubleshoot: %v", err)
	}
	if result["dns_resolved"] != true || result["configuration"] != "kitchen.yaml" {
		t.Fatalf("result = %+v", result)
	}
}

func TestDeviceBuilderSilentGreetingIsConnectionError(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ws" {
			t.Errorf("unexpected request to %s", r.URL.Path)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, _, _ = conn.ReadMessage() // never greet
	}))
	t.Cleanup(server.Close)

	c := NewESPHomeClient(server.URL, "")
	c.Timeout = 200 * time.Millisecond
	_, err := c.ReadConfig("kitchen.yaml")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != ErrCodeConnectionError {
		t.Fatalf("err = %v, want CONNECTION_ERROR", err)
	}
}

func TestLegacyDashboardRejected(t *testing.T) {
	for name, handler := range map[string]http.HandlerFunc{
		// The legacy dashboard answers /ws with 404.
		"404": http.NotFound,
		// A web app that answers every path with a page.
		"html": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(w, "<!doctype html>")
		},
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(handler)
			defer server.Close()
			c := NewESPHomeClient(server.URL, "")
			if _, err := c.ReadConfig("kitchen.yaml"); err == nil || !strings.Contains(err.Error(), "legacy ESPHome dashboard is no longer supported") {
				t.Fatalf("ReadConfig: err = %v, want the legacy dashboard error", err)
			}
			if _, err := c.GetJSONConfig("kitchen.yaml"); err == nil || !strings.Contains(err.Error(), "legacy ESPHome dashboard is no longer supported") {
				t.Fatalf("GetJSONConfig: err = %v, want the legacy dashboard error", err)
			}
		})
	}
}

func TestDeviceBuilderUnavailable(t *testing.T) {
	for status, code := range map[int]string{
		http.StatusUnauthorized:       ErrCodeAuthenticationError,
		http.StatusBadGateway:         ErrCodeConnectionError,
		http.StatusServiceUnavailable: ErrCodeConnectionError,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
			defer server.Close()
			_, err := NewESPHomeClient(server.URL, "").ReadConfig("kitchen.yaml")
			apiErr, ok := err.(*APIError)
			if !ok || apiErr.Code != code || strings.Contains(apiErr.Message, "legacy") {
				t.Fatalf("err = %#v, want code %s without the legacy dashboard message", err, code)
			}
		})
	}
}

func TestLegacyDashboardWebSocketRejected(t *testing.T) {
	// A WebSocket server that does not greet with a ServerInfoMessage.
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.WriteJSON(map[string]any{"event": "line", "data": "hello"})
	}))
	defer server.Close()

	_, err := NewESPHomeClient(server.URL, "").ReadConfig("kitchen.yaml")
	if err == nil || !strings.Contains(err.Error(), "no longer supported") {
		t.Fatalf("err = %v, want the legacy dashboard error", err)
	}
}

func TestDeviceBuilderGetDevices(t *testing.T) {
	server, _ := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{
			"configured": []any{map[string]any{
				"name": "kitchen", "configuration": "kitchen.yaml", "target_platform": "ESP32",
				"runtime_state": map[string]any{"state": "online", "deployed_version": "2026.11.1"},
			}},
			"importable": []any{map[string]any{"name": "plug", "project_name": "acme.plug"}},
		}})
	}, nil)

	devices, err := NewESPHomeClient(server.URL, "").GetDevices()
	if err != nil {
		t.Fatalf("GetDevices: %v", err)
	}
	if len(devices.Configured) != 1 || len(devices.Importable) != 1 {
		t.Fatalf("devices = %+v", devices)
	}
	d := devices.Configured[0]
	if d.State != "online" || d.DeployedVersion != "2026.11.1" || d.TargetPlatform != "ESP32" {
		t.Fatalf("device = %+v", d)
	}
}

func TestDeviceBuilderGetBoards(t *testing.T) {
	server, received := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{
			"boards": []any{
				map[string]any{"id": "nodemcu-32s", "name": "NodeMCU-32S"},
				map[string]any{"id": "esp32dev", "name": "ESP32 Dev Board"},
			},
			"total": 2,
		}})
	}, nil)

	boards, err := NewESPHomeClient(server.URL, "").GetBoards("ESP32")
	if err != nil {
		t.Fatalf("GetBoards: %v", err)
	}
	if len(boards) != 2 || boards[0].ID != "esp32dev" {
		t.Fatalf("boards = %+v, want sorted by name", boards)
	}
	cmd := received()[0]
	if cmd.Command != "boards/get_boards" || cmd.Args["platform"] != "esp32" {
		t.Fatalf("command = %+v", cmd)
	}
}

func TestDeviceBuilderCreateConfig(t *testing.T) {
	server, received := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{"configuration": "demo.yaml"}})
	}, nil)
	c := NewESPHomeClient(server.URL, "")

	result, err := c.CreateConfig(ESPHomeCreateRequest{Name: "demo", Board: "esp32dev", SSID: "net", PSK: "secret"})
	if err != nil || result.Configuration != "demo.yaml" {
		t.Fatalf("CreateConfig = %+v, %v", result, err)
	}
	if _, err := c.CreateConfig(ESPHomeCreateRequest{Name: "uploaded", FileContent: []byte("esphome:\n  name: uploaded\n")}); err != nil {
		t.Fatalf("CreateConfig upload: %v", err)
	}

	cmds := received()
	board := cmds[0].Args
	if cmds[0].Command != "devices/create" || board["name"] != "demo" || board["board_id"] != "esp32dev" || board["ssid"] != "net" || board["psk"] != "secret" {
		t.Fatalf("board create = %+v", cmds[0])
	}
	upload := cmds[1].Args
	if upload["file_content"] != "esphome:\n  name: uploaded\n" {
		t.Fatalf("file_content = %#v, want the YAML as plain text", upload["file_content"])
	}
	if _, ok := upload["board_id"]; ok {
		t.Fatal("upload create must not send board_id")
	}
}

func TestDeviceBuilderImportConfig(t *testing.T) {
	server, received := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{"configuration": "plug.yaml", "warning": "kept despite a package error"}})
	}, nil)
	c := NewESPHomeClient(server.URL, "")

	result, err := c.ImportConfig(ESPHomeImportRequest{Name: "plug", ProjectName: "acme.plug", PackageImportURL: "github://acme/plug.yaml", Encryption: true})
	if err != nil || result.Configuration != "plug.yaml" || result.Warning == "" {
		t.Fatalf("ImportConfig = %+v, %v", result, err)
	}
	if _, err := c.ImportConfig(ESPHomeImportRequest{Name: "plug2", ProjectName: "acme.plug", PackageImportURL: "github://acme/plug.yaml"}); err != nil {
		t.Fatalf("ImportConfig: %v", err)
	}

	cmds := received()
	if cmds[0].Command != "devices/import" || cmds[0].Args["package_import_url"] != "github://acme/plug.yaml" || cmds[0].Args["encryption"] == nil {
		t.Fatalf("import = %+v", cmds[0])
	}
	if _, ok := cmds[1].Args["encryption"]; ok {
		t.Fatal("encryption must be omitted when not requested")
	}
}

func TestDeviceBuilderSerialPortsAndVersion(t *testing.T) {
	server, _ := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		switch cmd.Command {
		case "config/serial_ports":
			send(map[string]any{"message_id": cmd.MessageID, "result": []any{map[string]any{"port": "COM3", "desc": "USB Serial", "hint": "esp"}}})
		case "config/version":
			send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{"server_version": "1.0.0", "esphome_version": "2026.11.1"}})
		}
	}, nil)
	c := NewESPHomeClient(server.URL, "")

	ports, err := c.GetSerialPorts()
	if err != nil || len(ports) != 1 || ports[0].Port != "COM3" {
		t.Fatalf("GetSerialPorts = %+v, %v", ports, err)
	}
	version, err := c.GetVersion()
	if err != nil || version != "2026.11.1" {
		t.Fatalf("GetVersion = %q, %v", version, err)
	}
}

// jobServer answers firmware job commands: each queued job's id is its
// command, and following it streams one output line and the given result.
func jobServer(t *testing.T, results map[string]map[string]any) (*httptest.Server, func() []wsCommand) {
	return newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		switch cmd.Command {
		case "firmware/compile", "firmware/upload":
			send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{"job_id": cmd.Command, "status": "queued"}})
		case "firmware/follow_job":
			id, _ := cmd.Args["job_id"].(string)
			send(map[string]any{"message_id": cmd.MessageID, "event": "output", "data": id + " running\n"})
			send(map[string]any{"message_id": cmd.MessageID, "event": "result", "data": results[id]})
			send(map[string]any{"message_id": cmd.MessageID, "result": nil})
		}
	}, nil)
}

func TestDeviceBuilderBuildFollowsJob(t *testing.T) {
	server, received := jobServer(t, map[string]map[string]any{
		"firmware/compile": {"status": "completed", "exit_code": 0, "error": nil},
	})

	var events []ESPHomeStreamEvent
	code, err := NewESPHomeClient(server.URL, "").StreamCommand("build", "kitchen.yaml", "", func(e ESPHomeStreamEvent) {
		events = append(events, e)
	})
	if err != nil || code != 0 {
		t.Fatalf("StreamCommand: code=%d err=%v", code, err)
	}
	if len(events) != 2 || events[0].Data != "firmware/compile running\n" || events[1].Event != "exit" {
		t.Fatalf("events = %+v", events)
	}
	cmds := received()
	if cmds[0].Command != "firmware/compile" || cmds[1].Command != "firmware/follow_job" || cmds[1].Args["job_id"] != "firmware/compile" {
		t.Fatalf("commands = %+v", cmds)
	}
}

func TestDeviceBuilderFailedJobReportsError(t *testing.T) {
	server, _ := jobServer(t, map[string]map[string]any{
		"firmware/upload": {"status": "failed", "exit_code": nil, "error": "Device not reachable"},
	})

	var lines []string
	code, err := NewESPHomeClient(server.URL, "").StreamCommand("upload", "kitchen.yaml", "OTA", func(e ESPHomeStreamEvent) {
		if e.Event == "line" {
			lines = append(lines, e.Data)
		}
	})
	if err != nil || code != 1 {
		t.Fatalf("StreamCommand: code=%d err=%v, want code 1", code, err)
	}
	if len(lines) != 2 || lines[1] != "Device not reachable\n" {
		t.Fatalf("lines = %q", lines)
	}
}

func TestDeviceBuilderRunCompilesThenUploads(t *testing.T) {
	server, received := jobServer(t, map[string]map[string]any{
		"firmware/compile": {"status": "completed", "exit_code": 0},
		"firmware/upload":  {"status": "completed", "exit_code": 0},
	})

	exits := 0
	code, err := NewESPHomeClient(server.URL, "").StreamCommand("run", "kitchen.yaml", "OTA", func(e ESPHomeStreamEvent) {
		if e.Event == "exit" {
			exits++
		}
	})
	if err != nil || code != 0 || exits != 1 {
		t.Fatalf("StreamCommand: code=%d err=%v exits=%d", code, err, exits)
	}
	var queued []string
	for _, cmd := range received() {
		if cmd.Command != "firmware/follow_job" {
			queued = append(queued, cmd.Command)
		}
		if cmd.Command == "firmware/upload" && cmd.Args["port"] != "OTA" {
			t.Fatalf("upload port = %#v, want OTA", cmd.Args["port"])
		}
	}
	if strings.Join(queued, ",") != "firmware/compile,firmware/upload" {
		t.Fatalf("queued = %v", queued)
	}
}

func TestDeviceBuilderRunStopsWhenBuildFails(t *testing.T) {
	server, received := jobServer(t, map[string]map[string]any{
		"firmware/compile": {"status": "failed", "exit_code": 2},
	})

	code, err := NewESPHomeClient(server.URL, "").StreamCommand("run", "kitchen.yaml", "OTA", func(ESPHomeStreamEvent) {})
	if err != nil || code != 2 {
		t.Fatalf("StreamCommand: code=%d err=%v, want 2", code, err)
	}
	for _, cmd := range received() {
		if cmd.Command == "firmware/upload" {
			t.Fatal("upload must not run after a failed build")
		}
	}
}

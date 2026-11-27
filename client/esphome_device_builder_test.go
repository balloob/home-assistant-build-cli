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
// plus any extra REST/legacy routes. It records every command received.
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
	code, err := NewESPHomeClient(server.URL, "").StreamCommand("/logs", map[string]any{"type": "spawn", "configuration": "kitchen.yaml", "port": "OTA"}, func(e ESPHomeStreamEvent) {
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

func TestDeviceBuilderRunUsesCompileThenUpload(t *testing.T) {
	upgrader := websocket.Upgrader{}
	var spawns []map[string]any
	legacy := func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		var spawn map[string]any
		_ = conn.ReadJSON(&spawn)
		spawn["path"] = r.URL.Path
		spawns = append(spawns, spawn)
		_ = conn.WriteJSON(map[string]any{"event": "line", "data": r.URL.Path + " ok\n"})
		_ = conn.WriteJSON(map[string]any{"event": "exit", "code": 0})
	}
	server, _ := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{"pong": true}})
	}, map[string]http.HandlerFunc{"/compile": legacy, "/upload": legacy})

	var events []ESPHomeStreamEvent
	code, err := NewESPHomeClient(server.URL, "").StreamCommand("/run", map[string]any{"type": "spawn", "configuration": "kitchen.yaml", "port": "OTA"}, func(e ESPHomeStreamEvent) {
		events = append(events, e)
	})
	if err != nil || code != 0 {
		t.Fatalf("StreamCommand: code=%d err=%v", code, err)
	}
	if len(spawns) != 2 || spawns[0]["path"] != "/compile" || spawns[1]["path"] != "/upload" || spawns[1]["port"] != "OTA" {
		t.Fatalf("spawns = %+v", spawns)
	}
	exits := 0
	for _, e := range events {
		if e.Event == "exit" {
			exits++
		}
	}
	if exits != 1 || len(events) != 3 {
		t.Fatalf("events = %+v, want two lines and one exit", events)
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

func TestClassicDashboardFallback(t *testing.T) {
	upgrader := websocket.Upgrader{}
	wsCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsCalls++
		http.NotFound(w, r)
	})
	mux.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "esphome:\n  name: kitchen\n")
	})
	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		var spawn map[string]any
		_ = conn.ReadJSON(&spawn)
		_ = conn.WriteJSON(map[string]any{"event": "line", "data": "classic log\n"})
		_ = conn.WriteJSON(map[string]any{"event": "exit", "code": 0})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewESPHomeClient(server.URL, "")
	content, err := c.ReadConfig("kitchen.yaml")
	if err != nil || content != "esphome:\n  name: kitchen\n" {
		t.Fatalf("ReadConfig = %q, %v", content, err)
	}
	var lines []string
	code, err := c.StreamCommand("/logs", map[string]any{"type": "spawn", "configuration": "kitchen.yaml", "port": "OTA"}, func(e ESPHomeStreamEvent) {
		if e.Event == "line" {
			lines = append(lines, e.Data)
		}
	})
	if err != nil || code != 0 || len(lines) != 1 || lines[0] != "classic log\n" {
		t.Fatalf("StreamCommand: code=%d err=%v lines=%q", code, err, lines)
	}
	if wsCalls != 1 {
		t.Fatalf("/ws probed %d times, want 1 (the result is cached)", wsCalls)
	}
}

func TestClassicDashboardWriteConfigIfUnchangedConflict(t *testing.T) {
	wrote := false
	mux := http.NewServeMux()
	mux.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			wrote = true
		}
		_, _ = io.WriteString(w, "other: 2\n")
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	err := NewESPHomeClient(server.URL, "").WriteConfigIfUnchanged("kitchen.yaml", "new: 1\n", "old: 1\n")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != ErrCodeValidationError {
		t.Fatalf("err = %v, want VALIDATION_ERROR", err)
	}
	if wrote {
		t.Fatal("config was written despite the conflict")
	}
}

func TestReadConfigRejectsHTMLPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, "<!doctype html><title>ESPHome Device Builder</title>")
	}))
	defer server.Close()

	_, err := NewESPHomeClient(server.URL, "").ReadConfig("kitchen.yaml")
	if err == nil || !strings.Contains(err.Error(), "web page") {
		t.Fatalf("err = %v, want an HTML rejection", err)
	}
}

func TestDeviceBuilderSilentGreetingIsNotClassic(t *testing.T) {
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
	if c.legacyDashboard {
		t.Fatal("a silent /ws must not mark the server as the classic dashboard")
	}
}

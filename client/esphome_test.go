package client

import (
	"io"
	"net/http"
	"testing"
)

// pong answers every /ws command, which GetJSONConfig uses to confirm the
// server is the Device Builder before its REST call.
func pong(cmd wsCommand, send func(any)) {
	send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{"pong": true}})
}

func TestESPHomeGetJSONConfig(t *testing.T) {
	server, _ := newDeviceBuilderServer(t, pong, map[string]http.HandlerFunc{"/json-config": func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("configuration") != "kitchen.yaml" {
			t.Errorf("configuration = %q, want kitchen.yaml", r.URL.Query().Get("configuration"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"esphome":{"name":"kitchen"}}`)
	}})

	result, err := NewESPHomeClient(server.URL, "").GetJSONConfig("kitchen.yaml")
	if err != nil {
		t.Fatalf("GetJSONConfig: %v", err)
	}
	esphome, ok := result["esphome"].(map[string]any)
	if !ok || esphome["name"] != "kitchen" {
		t.Fatalf("result = %#v, want parsed esphome name", result)
	}
}

func TestESPHomeGetJSONConfigValidationError(t *testing.T) {
	raw := "Error while reading config: required option 'ssid' is missing\n[wifi]\nssid: [source living-room.yaml:12:7]"
	server, _ := newDeviceBuilderServer(t, pong, map[string]http.HandlerFunc{"/json-config": func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, raw)
	}})

	_, err := NewESPHomeClient(server.URL, "").GetJSONConfig("living-room.yaml")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Code != ErrCodeValidationError {
		t.Fatalf("Code = %q, want %q", apiErr.Code, ErrCodeValidationError)
	}
	if apiErr.Details["configuration"] != "living-room.yaml" {
		t.Fatalf("configuration detail = %#v", apiErr.Details["configuration"])
	}
	if apiErr.Details["component"] != "wifi" {
		t.Fatalf("component detail = %#v, want wifi", apiErr.Details["component"])
	}
	if apiErr.Details["line"] != 12 || apiErr.Details["column"] != 7 {
		t.Fatalf("line/column = %#v/%#v, want 12/7", apiErr.Details["line"], apiErr.Details["column"])
	}
}

func TestESPHomeGetInfo(t *testing.T) {
	server, _ := newDeviceBuilderServer(t, func(cmd wsCommand, send func(any)) {
		if cmd.Command != "devices/list" {
			pong(cmd, send)
			return
		}
		send(map[string]any{"message_id": cmd.MessageID, "result": map[string]any{
			"configured": []any{map[string]any{"name": "plug-test", "friendly_name": "Plug Test", "configuration": "plug-test.yaml", "current_version": "2026.11.1", "address": "plug-test.local", "target_platform": "ESP32C3"}},
			"importable": []any{},
		}})
	}, map[string]http.HandlerFunc{"/json-config": func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"esphome":{"name":"plug-test","friendly_name":"Plug Test","build_path":"build/plug-test"},"esp32":{"variant":"ESP32C3","framework":{"type":"esp-idf"}},"switch":[{"platform":"gpio"}]}`)
	}})

	info, err := NewESPHomeClient(server.URL, "").GetInfo("plug-test.yaml")
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info["source"] != "devices+json_config" {
		t.Fatalf("source = %#v, want devices+json_config", info["source"])
	}
	if info["name"] != "plug-test" || info["address"] != "plug-test.local" {
		t.Fatalf("info = %#v", info)
	}
	if info["esp_platform"] != "ESP32C3" || info["core_platform"] != "esp32" || info["framework"] != "esp-idf" {
		t.Fatalf("platform info = %#v/%#v/%#v", info["esp_platform"], info["core_platform"], info["framework"])
	}
	if info["esphome_version"] != "2026.11.1" {
		t.Fatalf("esphome_version = %#v", info["esphome_version"])
	}
	loadedPlatforms, ok := info["loaded_platforms"].([]string)
	if !ok || len(loadedPlatforms) != 1 || loadedPlatforms[0] != "switch/gpio" {
		t.Fatalf("loaded_platforms = %#v, want [switch/gpio]", info["loaded_platforms"])
	}
}

package client

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestESPHomeGetBoards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/boards/esp32" {
			t.Fatalf("path = %q, want /boards/esp32", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[{"items":{"nodemcu-32s":"NodeMCU-32S","esp32dev":"ESP32 Dev Board"}}]`)
	}))
	defer server.Close()

	client := NewESPHomeClient(server.URL, "")
	boards, err := client.GetBoards("ESP32")
	if err != nil {
		t.Fatalf("GetBoards: %v", err)
	}
	if len(boards) != 2 {
		t.Fatalf("len(boards) = %d, want 2", len(boards))
	}
	if boards[0].ID != "esp32dev" {
		t.Fatalf("boards[0].ID = %q, want esp32dev", boards[0].ID)
	}
}

func TestESPHomeCreateConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wizard" {
			t.Fatalf("path = %q, want /wizard", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["type"] != "upload" {
			t.Fatalf("type = %#v, want upload", payload["type"])
		}
		decoded, err := base64.StdEncoding.DecodeString(payload["file_content"].(string))
		if err != nil {
			t.Fatalf("DecodeString: %v", err)
		}
		if string(decoded) != "esphome:\n  name: kitchen\n" {
			t.Fatalf("decoded file = %q", string(decoded))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"configuration":"kitchen.yaml"}`)
	}))
	defer server.Close()

	client := NewESPHomeClient(server.URL, "")
	result, err := client.CreateConfig(ESPHomeCreateRequest{
		Type:        "upload",
		Name:        "Kitchen",
		FileContent: []byte("esphome:\n  name: kitchen\n"),
	})
	if err != nil {
		t.Fatalf("CreateConfig: %v", err)
	}
	if result.Configuration != "kitchen.yaml" {
		t.Fatalf("Configuration = %q, want kitchen.yaml", result.Configuration)
	}
}

func TestESPHomeImportConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/import" {
			t.Fatalf("path = %q, want /import", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"configuration":"plug.yaml"}`)
	}))
	defer server.Close()

	client := NewESPHomeClient(server.URL, "")
	result, err := client.ImportConfig(ESPHomeImportRequest{
		Name:             "plug",
		ProjectName:      "esphome.demo",
		PackageImportURL: "https://example.com/demo.yaml",
	})
	if err != nil {
		t.Fatalf("ImportConfig: %v", err)
	}
	if result.Configuration != "plug.yaml" {
		t.Fatalf("Configuration = %q, want plug.yaml", result.Configuration)
	}
}

func TestESPHomeGetJSONConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json-config" {
			t.Fatalf("path = %q, want /json-config", r.URL.Path)
		}
		if r.URL.Query().Get("configuration") != "kitchen.yaml" {
			t.Fatalf("configuration = %q, want kitchen.yaml", r.URL.Query().Get("configuration"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"esphome":{"name":"kitchen"}}`)
	}))
	defer server.Close()

	client := NewESPHomeClient(server.URL, "")
	result, err := client.GetJSONConfig("kitchen.yaml")
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, raw)
	}))
	defer server.Close()

	client := NewESPHomeClient(server.URL, "")
	_, err := client.GetJSONConfig("living-room.yaml")
	if err == nil {
		t.Fatal("expected validation error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %T, want *APIError", err)
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
	if apiErr.Details["line"] != 12 {
		t.Fatalf("line detail = %#v, want 12", apiErr.Details["line"])
	}
	if apiErr.Details["column"] != 7 {
		t.Fatalf("column detail = %#v, want 7", apiErr.Details["column"])
	}
}

func TestESPHomeGetSerialPorts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/serial-ports" {
			t.Fatalf("path = %q, want /serial-ports", r.URL.Path)
		}
		_, _ = io.WriteString(w, `[{"port":"COM3","desc":"USB Serial"}]`)
	}))
	defer server.Close()

	client := NewESPHomeClient(server.URL, "")
	ports, err := client.GetSerialPorts()
	if err != nil {
		t.Fatalf("GetSerialPorts: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("len(ports) = %d, want 1", len(ports))
	}
	if ports[0].Port != "COM3" {
		t.Fatalf("ports[0].Port = %q, want COM3", ports[0].Port)
	}
}

func TestESPHomeGetInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/info" {
			t.Fatalf("path = %q, want /info", r.URL.Path)
		}
		if r.URL.Query().Get("configuration") != "kitchen.yaml" {
			t.Fatalf("configuration = %q, want kitchen.yaml", r.URL.Query().Get("configuration"))
		}
		_, _ = io.WriteString(w, `{"name":"kitchen","target_platform":"ESP32"}`)
	}))
	defer server.Close()

	client := NewESPHomeClient(server.URL, "")
	info, err := client.GetInfo("kitchen.yaml")
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info["name"] != "kitchen" {
		t.Fatalf("name = %#v, want kitchen", info["name"])
	}
}

func TestESPHomeGetInfoFallsBackToJSONConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/info":
			w.WriteHeader(http.StatusNotFound)
		case "/json-config":
			_, _ = io.WriteString(w, `{"esphome":{"name":"plug-test","friendly_name":"Plug Test","build_path":"build/plug-test"},"esp32":{"variant":"ESP32C3","framework":{"type":"esp-idf"}},"switch":[{"platform":"gpio"}]}`)
		case "/devices":
			_, _ = io.WriteString(w, `{"configured":[{"name":"plug-test","friendly_name":"Plug Test","configuration":"plug-test.yaml","current_version":"2026.2.4","path":"/config/esphome/plug-test.yaml","address":"plug-test.local","target_platform":"ESP32C3"}],"importable":[]}`)
		case "/version":
			_, _ = io.WriteString(w, `{"version":"2026.2.4"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewESPHomeClient(server.URL, "")
	info, err := client.GetInfo("plug-test.yaml")
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if info["source"] != "devices+json_config" {
		t.Fatalf("source = %#v, want devices+json_config", info["source"])
	}
	if info["name"] != "plug-test" {
		t.Fatalf("name = %#v, want plug-test", info["name"])
	}
	if info["esp_platform"] != "ESP32C3" {
		t.Fatalf("esp_platform = %#v, want ESP32C3", info["esp_platform"])
	}
	if info["core_platform"] != "esp32" {
		t.Fatalf("core_platform = %#v, want esp32", info["core_platform"])
	}
	if info["framework"] != "esp-idf" {
		t.Fatalf("framework = %#v, want esp-idf", info["framework"])
	}
	loadedPlatforms, ok := info["loaded_platforms"].([]string)
	if !ok || len(loadedPlatforms) != 1 || loadedPlatforms[0] != "switch/gpio" {
		t.Fatalf("loaded_platforms = %#v, want [switch/gpio]", info["loaded_platforms"])
	}
}

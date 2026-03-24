package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type serverState struct {
	mu      sync.RWMutex
	configs map[string]string
}

func main() {
	port := flag.Int("port", 16123, "Port to bind the mock ESPHome dashboard")
	flag.Parse()

	state := &serverState{configs: map[string]string{}}

	mux := http.NewServeMux()
	mux.HandleFunc("/devices", state.handleDevices)
	mux.HandleFunc("/ping", state.handlePing)
	mux.HandleFunc("/boards/", state.handleBoards)
	mux.HandleFunc("/wizard", state.handleWizard)
	mux.HandleFunc("/import", state.handleImport)
	mux.HandleFunc("/edit", state.handleEdit)
	mux.HandleFunc("/json-config", state.handleJSONConfig)
	mux.HandleFunc("/serial-ports", state.handleSerialPorts)
	mux.HandleFunc("/info", state.handleInfo)
	mux.HandleFunc("/version", state.handleVersion)

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	log.Printf("mock esphome dashboard listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *serverState) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	configured := make([]map[string]any, 0, len(s.configs))
	for configuration := range s.configs {
		configured = append(configured, map[string]any{
			"name":                strings.TrimSuffix(configuration, ".yaml"),
			"friendly_name":       strings.TrimSuffix(configuration, ".yaml"),
			"configuration":       configuration,
			"loaded_integrations": []string{"api", "wifi"},
			"target_platform":     "ESP32",
		})
	}
	s.mu.RUnlock()

	writeJSON(w, map[string]any{"configured": configured, "importable": []any{}})
}

func (s *serverState) handlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	status := map[string]*bool{}
	for configuration := range s.configs {
		online := true
		status[configuration] = &online
	}
	s.mu.RUnlock()

	writeJSON(w, status)
}

func (s *serverState) handleBoards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, []map[string]any{
		{
			"items": map[string]string{
				"esp32dev":    "ESP32 Dev Board",
				"nodemcu-32s": "NodeMCU-32S",
				"esp01_1m":    "Generic ESP8266 Module",
			},
		},
	})
}

func (s *serverState) handleWizard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := map[string]any{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name, _ := payload["name"].(string)
	if strings.TrimSpace(name) == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		writeJSON(w, map[string]any{"error": "name is required"})
		return
	}

	typeName, _ := payload["type"].(string)
	if typeName == "" {
		typeName = "basic"
	}

	configuration := slugify(name) + ".yaml"
	content, err := buildWizardContent(typeName, payload)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}

	s.mu.Lock()
	s.configs[configuration] = content
	s.mu.Unlock()

	writeJSON(w, map[string]any{"configuration": configuration})
}

func (s *serverState) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := map[string]any{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name, _ := payload["name"].(string)
	if strings.TrimSpace(name) == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		writeJSON(w, map[string]any{"error": "name is required"})
		return
	}

	configuration := slugify(name) + ".yaml"
	content := fmt.Sprintf("esphome:\n  name: %s\n\n# imported from package\n", slugify(name))

	s.mu.Lock()
	s.configs[configuration] = content
	s.mu.Unlock()

	writeJSON(w, map[string]any{"configuration": configuration})
}

func (s *serverState) handleEdit(w http.ResponseWriter, r *http.Request) {
	configuration := r.URL.Query().Get("configuration")
	if configuration == "" {
		http.Error(w, "missing configuration", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		content, ok := s.configs[configuration]
		s.mu.RUnlock()
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(content))
	case http.MethodPost:
		body, err := ioReadAll(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		s.configs[configuration] = string(body)
		s.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *serverState) handleJSONConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	configuration := r.URL.Query().Get("configuration")
	s.mu.RLock()
	content, ok := s.configs[configuration]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if strings.Contains(content, "invalid:") {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("Error while reading config\n[source " + configuration + ":3:3]"))
		return
	}

	writeJSON(w, map[string]any{
		"esphome": map[string]any{
			"name": strings.TrimSuffix(configuration, ".yaml"),
		},
	})
}

func (s *serverState) handleSerialPorts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, []map[string]any{
		{"port": "COM3", "desc": "USB Serial"},
		{"port": "OTA", "desc": "Over-The-Air"},
	})
}

func (s *serverState) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	configuration := r.URL.Query().Get("configuration")
	s.mu.RLock()
	_, ok := s.configs[configuration]
	s.mu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{
		"name":            strings.TrimSuffix(configuration, ".yaml"),
		"friendly_name":   strings.TrimSuffix(configuration, ".yaml"),
		"target_platform": "ESP32",
	})
}

func (s *serverState) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]any{"version": "2026.3.0-mock"})
}

func buildWizardContent(typeName string, payload map[string]any) (string, error) {
	name := slugify(anyString(payload["name"]))
	platform := strings.ToLower(anyString(payload["platform"]))
	board := anyString(payload["board"])
	ssid := anyString(payload["ssid"])
	psk := anyString(payload["psk"])

	switch typeName {
	case "empty":
		return "", nil
	case "upload":
		encoded := anyString(payload["file_content"])
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return "", fmt.Errorf("invalid file_content")
		}
		return string(decoded), nil
	case "basic":
		if board == "" {
			return "", fmt.Errorf("board is required")
		}
		if platform == "" {
			platform = "esp32"
		}

		platformBlock := "esp32:\n  board: " + board + "\n"
		if platform == "esp8266" {
			platformBlock = "esp8266:\n  board: " + board + "\n"
		}

		if ssid == "" {
			ssid = "test-ssid"
		}
		if psk == "" {
			psk = "test-password"
		}

		return "esphome:\n  name: " + name + "\n\n" +
			platformBlock + "\n" +
			"logger:\n\n" +
			"api:\n  encryption:\n    key: mock_encryption_key\n\n" +
			"ota:\n  - platform: esphome\n    password: mock_ota_password\n\n" +
			"wifi:\n  ssid: \"" + ssid + "\"\n  password: \"" + psk + "\"\n", nil
	default:
		return "", fmt.Errorf("unsupported wizard type")
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func anyString(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func slugify(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	lower = strings.ReplaceAll(lower, "_", "-")
	lower = strings.ReplaceAll(lower, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9\-]+`)
	lower = re.ReplaceAllString(lower, "")
	lower = strings.Trim(lower, "-")
	if lower == "" {
		lower = "device"
	}
	return lower
}

func ioReadAll(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

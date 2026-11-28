// Command mock_esphome_device_builder is a fake ESPHome Device Builder for the
// hab integration tests. It speaks the /ws protocol for the commands hab uses
// and serves the /json-config and /version REST endpoints. Paths under
// /legacy/ answer 404, like the legacy ESPHome dashboard does for /ws.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type serverState struct {
	mu      sync.RWMutex
	configs map[string]string
}

type command struct {
	Command   string         `json:"command"`
	MessageID string         `json:"message_id"`
	Args      map[string]any `json:"args"`
}

// commandError is a /ws error reply.
type commandError struct {
	code    string
	details string
}

func (e *commandError) Error() string { return e.details }

var boards = []map[string]any{
	{"id": "esp32dev", "name": "ESP32 Dev Board", "esphome": map[string]any{"platform": "esp32", "board": "esp32dev"}},
	{"id": "nodemcu-32s", "name": "NodeMCU-32S", "esphome": map[string]any{"platform": "esp32", "board": "nodemcu-32s"}},
	{"id": "esp01_1m", "name": "Generic ESP8266 Module", "esphome": map[string]any{"platform": "esp8266", "board": "esp01_1m"}},
}

func main() {
	port := flag.Int("port", 16123, "Port to bind the mock ESPHome Device Builder")
	flag.Parse()

	state := &serverState{configs: map[string]string{}}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", state.handleWS)
	mux.HandleFunc("/json-config", state.handleJSONConfig)
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"version": "2026.11.1-mock"})
	})
	mux.HandleFunc("/legacy/", http.NotFound)

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	log.Printf("mock esphome device builder listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *serverState) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = conn.WriteJSON(map[string]any{"server_version": "mock", "esphome_version": "2026.11.1-mock", "port": 6052, "requires_auth": false})
	for {
		var cmd command
		if err := conn.ReadJSON(&cmd); err != nil {
			return
		}
		event := func(name string, data any) {
			_ = conn.WriteJSON(map[string]any{"message_id": cmd.MessageID, "event": name, "data": data})
		}
		result, err := s.dispatch(cmd, event)
		if cmdErr, ok := err.(*commandError); ok {
			_ = conn.WriteJSON(map[string]any{"message_id": cmd.MessageID, "error_code": cmdErr.code, "details": cmdErr.details})
			continue
		}
		_ = conn.WriteJSON(map[string]any{"message_id": cmd.MessageID, "result": result})
	}
}

func (s *serverState) dispatch(cmd command, event func(string, any)) (any, error) {
	configuration := anyString(cmd.Args["configuration"])
	switch cmd.Command {
	case "ping":
		return map[string]any{"pong": true}, nil
	case "config/version":
		return map[string]any{"server_version": "mock", "esphome_version": "2026.11.1-mock"}, nil
	case "config/serial_ports":
		return []map[string]any{{"port": "COM3", "desc": "USB Serial", "hint": "bridge"}}, nil
	case "devices/list":
		return s.devices(), nil
	case "boards/get_boards":
		platform := anyString(cmd.Args["platform"])
		matched := []map[string]any{}
		for _, b := range boards {
			if platform == "" || b["esphome"].(map[string]any)["platform"] == platform {
				matched = append(matched, b)
			}
		}
		return map[string]any{"boards": matched, "total": len(matched), "offset": 0, "limit": len(matched)}, nil
	case "devices/create":
		return s.create(cmd.Args)
	case "devices/import":
		name := slugify(anyString(cmd.Args["name"]))
		configuration := name + ".yaml"
		s.mu.Lock()
		s.configs[configuration] = fmt.Sprintf("esphome:\n  name: %s\n\n# imported from package\n", name)
		s.mu.Unlock()
		return map[string]any{"configuration": configuration}, nil
	case "devices/get_config":
		content, ok := s.config(configuration)
		if !ok {
			return nil, &commandError{"not_found", "Configuration " + configuration + " not found"}
		}
		return content, nil
	case "devices/update_config":
		return nil, s.updateConfig(configuration, anyString(cmd.Args["content"]), cmd.Args["expected"])
	case "devices/troubleshoot":
		if _, ok := s.config(configuration); !ok {
			return nil, &commandError{"not_found", "Configuration " + configuration + " not found"}
		}
		return map[string]any{"configuration": configuration, "address": strings.TrimSuffix(configuration, ".yaml") + ".local", "icmp_available": true, "zeroconf_running": true, "dns_resolved": true, "dns_addresses": []string{"192.0.2.10"}, "ping_attempted": true, "ping_rtt_ms": 3.2}, nil
	case "devices/validate", "devices/logs":
		content, ok := s.config(configuration)
		if !ok {
			return nil, &commandError{"not_found", "Configuration " + configuration + " not found"}
		}
		if cmd.Command == "devices/validate" && strings.Contains(content, "invalid:") {
			event("output", "Error while reading config")
			event("output", "[source "+configuration+":3:3]")
			event("result", map[string]any{"success": false, "code": 2})
			return nil, nil
		}
		event("output", "INFO "+cmd.Command+" "+configuration)
		event("result", map[string]any{"success": true, "code": 0})
		return nil, nil
	case "firmware/compile", "firmware/upload":
		if _, ok := s.config(configuration); !ok {
			return nil, &commandError{"not_found", "Configuration " + configuration + " not found"}
		}
		return map[string]any{"job_id": cmd.Command + ":" + configuration, "configuration": configuration, "status": "queued"}, nil
	case "firmware/follow_job":
		event("output", "INFO "+anyString(cmd.Args["job_id"])+"\n")
		event("result", map[string]any{"status": "completed", "exit_code": 0, "error": nil})
		return nil, nil
	default:
		return nil, &commandError{"unknown_command", "Unknown command: " + cmd.Command}
	}
}

func (s *serverState) devices() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	configured := make([]map[string]any, 0, len(s.configs))
	for configuration := range s.configs {
		name := strings.TrimSuffix(configuration, ".yaml")
		configured = append(configured, map[string]any{
			"name":                name,
			"friendly_name":       name,
			"configuration":       configuration,
			"loaded_integrations": []string{"api", "wifi"},
			"target_platform":     "ESP32",
			"runtime_state":       map[string]any{"state": "online", "deployed_version": "2026.11.1-mock"},
		})
	}
	return map[string]any{"configured": configured, "importable": []any{}}
}

func (s *serverState) config(configuration string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	content, ok := s.configs[configuration]
	return content, ok
}

func (s *serverState) updateConfig(configuration, content string, expected any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if expected != nil {
		current, ok := s.configs[configuration]
		if !ok {
			return &commandError{"not_found", "Configuration " + configuration + " not found"}
		}
		if strings.TrimRight(current, "\n") != strings.TrimRight(anyString(expected), "\n") {
			return &commandError{"precondition_failed", configuration + " changed since it was read"}
		}
	}
	s.configs[configuration] = content
	return nil
}

func (s *serverState) create(args map[string]any) (any, error) {
	name := slugify(anyString(args["name"]))
	if anyString(args["name"]) == "" {
		return nil, &commandError{"invalid_args", "name is required"}
	}
	configuration := name + ".yaml"

	content := anyString(args["file_content"])
	if content == "" {
		platformBlock := "esp32:\n  board: esp32dev\n"
		if boardID := anyString(args["board_id"]); boardID != "" {
			board := findBoard(boardID)
			if board == nil {
				return nil, &commandError{"invalid_args", "Unknown board: " + boardID}
			}
			esphome := board["esphome"].(map[string]any)
			platformBlock = fmt.Sprintf("%s:\n  board: %s\n", esphome["platform"], esphome["board"])
		}
		content = "esphome:\n  name: " + name + "\n\n" +
			platformBlock + "\n" +
			"logger:\n\n" +
			"api:\n  encryption:\n    key: mock_encryption_key\n\n" +
			"ota:\n  - platform: esphome\n    password: mock_ota_password\n\n" +
			"wifi:\n  ssid: !secret wifi_ssid\n  password: !secret wifi_password\n"
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.configs[configuration]; exists && args["overwrite"] != true {
		return nil, &commandError{"already_exists", configuration + " already exists"}
	}
	s.configs[configuration] = content
	return map[string]any{"configuration": configuration}, nil
}

func findBoard(id string) map[string]any {
	for _, b := range boards {
		if b["id"] == id {
			return b
		}
	}
	return nil
}

func (s *serverState) handleJSONConfig(w http.ResponseWriter, r *http.Request) {
	configuration := r.URL.Query().Get("configuration")
	content, ok := s.config(configuration)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if strings.Contains(content, "invalid:") {
		// Like the Device Builder, keep the details out of the body;
		// devices/validate streams them.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"Configuration is invalid"}`))
		return
	}

	writeJSON(w, map[string]any{
		"esphome": map[string]any{
			"name": strings.TrimSuffix(configuration, ".yaml"),
		},
	})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func anyString(value any) string {
	text, _ := value.(string)
	return text
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

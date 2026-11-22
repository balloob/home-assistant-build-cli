package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func appTestServer(t *testing.T, result any, failure map[string]any) (<-chan map[string]any, *atomic.Int32) {
	t.Helper()
	requests := make(chan map[string]any, 1)
	connections := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connections.Add(1)
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				t.Error(err)
			}
		}()
		if err := conn.WriteJSON(map[string]any{"type": "auth_required"}); err != nil {
			t.Error(err)
			return
		}
		var message map[string]any
		if err := conn.ReadJSON(&message); err != nil {
			t.Error(err)
			return
		}
		if message["access_token"] != "app-test-token" {
			t.Error("unexpected authentication")
		}
		if err := conn.WriteJSON(map[string]any{"type": "auth_ok"}); err != nil {
			t.Error(err)
			return
		}
		_, raw, err := conn.ReadMessage()
		if err != nil {
			t.Error(err)
			return
		}
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.UseNumber()
		if err := decoder.Decode(&message); err != nil {
			t.Error(err)
			return
		}
		requests <- message
		response := map[string]any{"id": message["id"], "type": "result", "success": failure == nil, "result": result}
		if failure != nil {
			response["error"] = failure
		}
		if err := conn.WriteJSON(response); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("HAB_URL", server.URL)
	t.Setenv("HAB_TOKEN", "app-test-token")
	previous := getAuthManager()
	cachedAuthMgr = auth.NewManager(t.TempDir())
	t.Cleanup(func() { cachedAuthMgr = previous })
	return requests, connections
}

func executeAppTest(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = file
	t.Cleanup(func() { os.Stdout = previous })
	oldText := viper.Get("text")
	viper.Set("text", false)
	defer viper.Set("text", oldText)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	err = cmd.Execute()
	os.Stdout = previous
	if closeErr := file.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	raw, readErr := os.ReadFile(file.Name())
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(raw), err
}

func TestAppRequests(t *testing.T) {
	for _, tc := range []struct {
		name, method, endpoint string
		command                func() *cobra.Command
		args                   []string
		data, result           any
	}{
		{"installed", "get", "/addons", newAppListCommand, nil, nil, map[string]any{"addons": []any{}}},
		{"store", "get", "/store", newAppListCommand, []string{"--available"}, nil, map[string]any{"addons": []any{}}},
		{"get", "get", "/addons/core_mosquitto/info", newAppGetCommand, []string{"core_mosquitto"}, nil, map[string]any{"slug": "core_mosquitto"}},
		{"store_get", "get", "/store/addons/core_mosquitto", newAppGetCommand, []string{"core_mosquitto", "--available"}, nil, map[string]any{"slug": "core_mosquitto"}},
		{"configure", "post", "/addons/core_mosquitto/options", newAppConfigureCommand, []string{"core_mosquitto", "--data", `{"options":{"id":9007199254740993},"auto_update":false,"network":null}`}, map[string]any{"options": map[string]any{"id": json.Number("9007199254740993")}, "auto_update": false, "network": nil}, nil},
		{"configure_yaml", "post", "/addons/core_mosquitto/options", newAppConfigureCommand, []string{"core_mosquitto", "--data", "boot: auto\nwatchdog: true"}, map[string]any{"boot": "auto", "watchdog": true}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests, _ := appTestServer(t, tc.result, nil)
			out, err := executeAppTest(t, tc.command(), tc.args...)
			if err != nil {
				t.Fatal(err)
			}
			request := <-requests
			if request["type"] != "supervisor/api" || request["method"] != tc.method || request["endpoint"] != tc.endpoint || !reflect.DeepEqual(request["data"], tc.data) {
				t.Fatalf("unexpected request: %#v", request)
			}
			if timeout, exists := request["timeout"]; !exists || timeout != nil {
				t.Fatal("Supervisor proxy timeout must be disabled")
			}
			if !strings.Contains(out, `"success": true`) {
				t.Fatalf("unexpected output: %s", out)
			}
		})
	}
}

func TestAppActions(t *testing.T) {
	for _, operation := range []string{"install", "update", "start", "stop", "restart", "uninstall"} {
		t.Run(operation, func(t *testing.T) {
			requests, _ := appTestServer(t, nil, nil)
			cmd := newAppActionCommand(operation)
			args := []string{"core_mosquitto"}
			if cmd.Flags().Lookup("force") != nil {
				args = append(args, "--force")
			}
			if _, err := executeAppTest(t, cmd, args...); err != nil {
				t.Fatal(err)
			}
			request := <-requests
			if request["endpoint"] != "/addons/core_mosquitto/"+operation || request["method"] != "post" {
				t.Fatalf("unexpected request: %v", request)
			}
			if operation == "uninstall" && !reflect.DeepEqual(request["data"], map[string]any{"remove_config": false}) {
				t.Fatal("uninstall must preserve configuration by default")
			}
		})
	}
	t.Run("remove_config", func(t *testing.T) {
		requests, _ := appTestServer(t, nil, nil)
		if _, err := executeAppTest(t, newAppActionCommand("uninstall"), "core_mosquitto", "--force", "--remove-config"); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual((<-requests)["data"], map[string]any{"remove_config": true}) {
			t.Fatal("explicit configuration removal was not sent")
		}
	})
}

func TestAppPlansAndConfirmationDoNotConnect(t *testing.T) {
	_, connections := appTestServer(t, nil, nil)
	for _, operation := range []string{"install", "update", "start", "stop", "restart", "uninstall"} {
		for _, flag := range []string{"--plan", "--dry-run"} {
			out, err := executeAppTest(t, newAppActionCommand(operation), "core_mosquitto", flag)
			if err != nil || !strings.Contains(out, `"mode": "plan"`) {
				t.Fatalf("%s %s: %s, %v", operation, flag, out, err)
			}
		}
	}
	for _, operation := range []string{"stop", "restart", "uninstall"} {
		if _, err := executeAppTest(t, newAppActionCommand(operation), "core_mosquitto"); err == nil {
			t.Fatalf("%s must require confirmation", operation)
		}
	}
	out, err := executeAppTest(t, newAppConfigureCommand(), "core_mosquitto", "--data", `{"options":{"custom":"private-value"}}`, "--plan")
	if err != nil || strings.Contains(out, "private-value") || !strings.Contains(out, `"mode": "plan"`) {
		t.Fatalf("unsafe configuration plan: %s, %v", out, err)
	}
	if connections.Load() != 0 {
		t.Fatal("preview or unconfirmed operation connected to Home Assistant")
	}
}

func TestAppInvalidInputDoesNotConnect(t *testing.T) {
	_, connections := appTestServer(t, nil, nil)
	for _, slug := range []string{"../supervisor", "core_mosquitto/info", "core_mosquitto?x=1", ""} {
		if _, err := executeAppTest(t, newAppGetCommand(), slug); err == nil {
			t.Errorf("accepted invalid slug %q", slug)
		}
	}
	for _, data := range []string{`{}`, `null`, `[]`, `{"options":[]}`, `{"boot":"invalid"}`, `{"boot":{}}`, `{"watchdog":"true"}`, `{"private-value":true}`} {
		_, err := executeAppTest(t, newAppConfigureCommand(), "core_mosquitto", "--data", data, "--plan")
		if err == nil || strings.Contains(err.Error(), "private-value") {
			t.Errorf("invalid configuration handling: %v", err)
		}
	}
	if _, err := executeAppTest(t, newAppActionCommand("install"), "core_mosquitto", "--timeout", "0s", "--plan"); err == nil {
		t.Fatal("accepted invalid timeout")
	}
	if connections.Load() != 0 {
		t.Fatal("invalid input connected to Home Assistant")
	}
}

func TestAppErrors(t *testing.T) {
	for _, code := range []string{"unknown_command", "unauthorized", "not_found", "unknown_error"} {
		t.Run(code, func(t *testing.T) {
			_, _ = appTestServer(t, nil, map[string]any{"code": code, "message": "Supervisor refused the request"})
			out, err := executeAppTest(t, newAppActionCommand("install"), "core_mosquitto")
			if err == nil || strings.Contains(out, `"success": true`) || !strings.Contains(err.Error(), "Supervisor refused") {
				t.Fatalf("failure reported as success: %s, %v", out, err)
			}
		})
	}
}

func TestAppUnknownOutcomeSurvivesErrorFormatting(t *testing.T) {
	for _, code := range []string{client.ErrCodeTimeout, client.ErrCodeConnectionError, client.ErrCodeCancelled} {
		t.Run(code, func(t *testing.T) {
			appTestServer(t, nil, map[string]any{"code": code, "message": "Operation interrupted"})
			out, err := executeAppTest(t, newAppActionCommand("install"), "core_mosquitto")
			if err == nil || strings.Contains(out, `"success": true`) {
				t.Fatalf("interrupted operation reported success: %s, %v", out, err)
			}
			gotCode, message, details := classifyError(err)
			if gotCode != code || !strings.Contains(message, "may still complete") || details["retryable"] != false || details["category"] != "unknown_outcome" {
				t.Fatalf("unsafe final error: %s, %s, %#v", gotCode, message, details)
			}
			if !reflect.DeepEqual(details["suggested_commands"], []string{"hab app list --json"}) {
				t.Fatalf("missing inspection command: %#v", details)
			}
		})
	}
}

func TestAppListFiltering(t *testing.T) {
	apps := map[string]any{"addons": []any{
		map[string]any{"slug": "core_mosquitto", "name": "Mosquitto", "description": "MQTT broker"},
		map[string]any{"slug": "core_ssh", "name": "Terminal", "description": "SSH server"},
	}}
	for search, count := range map[string]int{"": 2, "MQTT": 1, "core_ssh": 1, "terminal": 1, "absent": 0} {
		result, err := filterApps(apps, search)
		if err != nil || len(result) != count || result == nil {
			t.Fatalf("search %q: %v, %v", search, result, err)
		}
	}
	for _, result := range []any{nil, map[string]any{}, map[string]any{"addons": []any{false}}} {
		if _, err := filterApps(result, ""); err == nil {
			t.Fatalf("accepted malformed response: %v", result)
		}
	}
	for _, flag := range []string{"--count", "--brief"} {
		t.Run(flag, func(t *testing.T) {
			appTestServer(t, apps, nil)
			out, err := executeAppTest(t, newAppListCommand(), "--search", "MQTT", flag)
			if err != nil {
				t.Fatal(err)
			}
			var envelope struct{ Data json.RawMessage }
			if err := json.Unmarshal([]byte(out), &envelope); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(envelope.Data), "SSH") || strings.Contains(string(envelope.Data), "description") {
				t.Fatalf("unexpected filtered output: %s", envelope.Data)
			}
			if flag == "--count" {
				var count struct{ Count int }
				if err := json.Unmarshal(envelope.Data, &count); err != nil || count.Count != 1 {
					t.Fatalf("unexpected count: %s, %v", envelope.Data, err)
				}
			}
		})
	}
}

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func managementTestCommand(t *testing.T, path string) *cobra.Command {
	t.Helper()
	original, remaining, err := rootCmd.Find(strings.Fields(path))
	if err != nil || len(remaining) != 0 || !original.Runnable() {
		t.Fatalf("command %s: %v %v", path, remaining, err)
	}
	wrapper := &cobra.Command{Use: original.Use, Args: original.Args, RunE: func(_ *cobra.Command, args []string) error {
		original.SetContext(context.Background())
		return original.RunE(original, args)
	}}
	original.Flags().VisitAll(func(flag *pflag.Flag) {
		value, changed := flag.Value.String(), flag.Changed
		t.Cleanup(func() {
			if err := flag.Value.Set(value); err != nil {
				t.Error(err)
			}
			flag.Changed = changed
		})
	})
	wrapper.Flags().AddFlagSet(original.Flags())
	return wrapper
}

func TestManagementWireRequests(t *testing.T) {
	cases := []struct {
		path, args, data, requestType, endpoint, method string
		expected                                        map[string]any
	}{
		{path: "integration update", args: "entry", data: `{"title":"New"}`, requestType: "config_entries/update", expected: map[string]any{"entry_id": "entry", "title": "New"}},
		{path: "integration subentry delete", args: "entry sub", requestType: "config_entries/subentries/delete", expected: map[string]any{"entry_id": "entry", "subentry_id": "sub"}},
		{path: "assist pipeline get", args: "pipeline", requestType: "assist_pipeline/pipeline/get", expected: map[string]any{"pipeline_id": "pipeline"}},
		{path: "assist exposure set", data: `{"assistants":["conversation"],"entity_ids":["light.test"],"should_expose":false}`, requestType: "homeassistant/expose_entity", expected: map[string]any{"assistants": []any{"conversation"}, "entity_ids": []any{"light.test"}, "should_expose": false}},
		{path: "statistics adjust-sum", args: "sensor.test", data: `{"start_time":"2026-01-01T00:00:00Z","adjustment":9007199254740993,"adjustment_unit_of_measurement":null}`, requestType: "recorder/adjust_sum_statistics", expected: map[string]any{"statistic_id": "sensor.test", "start_time": "2026-01-01T00:00:00Z", "adjustment": json.Number("9007199254740993"), "adjustment_unit_of_measurement": nil}},
		{path: "repairs get", args: "domain issue", requestType: "repairs/get_issue_data", expected: map[string]any{"domain": "domain", "issue_id": "issue"}},
		{path: "user password", args: "user", data: `{"password":"secret"}`, requestType: "config/auth_provider/homeassistant/admin_change_password", expected: map[string]any{"user_id": "user", "password": "secret"}},
		{path: "application-credential create", data: `{"domain":"demo","client_id":"client","client_secret":"secret"}`, requestType: "application_credentials/create", expected: map[string]any{"domain": "demo", "client_id": "client", "client_secret": "secret"}},
		{path: "system config update", data: `{"external_url":null,"elevation":0}`, requestType: "config/core/update", expected: map[string]any{"external_url": nil, "elevation": json.Number("0")}},
		{path: "media search", args: "song", data: `{"media_content_id":"media-source://media_source/local"}`, requestType: "media_source/search_media", expected: map[string]any{"search_query": "song", "media_content_id": "media-source://media_source/local"}},
		{path: "app validate", args: "demo", data: `{}`, requestType: "supervisor/api", method: "post", endpoint: "/addons/demo/options/validate", expected: map[string]any{}},
		{path: "app repository add", args: "https://github.com/example/apps", requestType: "supervisor/api", method: "post", endpoint: "/store/repositories", expected: map[string]any{"repository": "https://github.com/example/apps"}},
		{path: "host datadisk move", args: "/dev/sda", requestType: "supervisor/api", method: "post", endpoint: "/os/datadisk/move", expected: map[string]any{"device": "/dev/sda"}},
		{path: "network interfaces update", args: "eth0", data: `{"ipv4":{"method":"auto"}}`, requestType: "supervisor/api", method: "post", endpoint: "/network/interface/eth0/update", expected: map[string]any{"ipv4": map[string]any{"method": "auto"}}},
		{path: "mount create", data: `{"name":"nas","type":"nfs","usage":"media","server":"nas.local","path":"/media"}`, requestType: "supervisor/api", method: "post", endpoint: "/mounts", expected: map[string]any{"name": "nas", "type": "nfs", "usage": "media", "server": "nas.local", "path": "/media"}},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			requests, _ := appTestServer(t, map[string]any{"accepted": true}, nil)
			cmd := managementTestCommand(t, tc.path)
			args := strings.Fields(tc.args)
			if tc.data != "" {
				args = append(args, "--data", tc.data)
			}
			if cmd.Flags().Lookup("force") != nil {
				args = append(args, "--force")
			}
			out, err := executeAppTest(t, cmd, args...)
			if err != nil {
				t.Fatal(err)
			}
			req := <-requests
			if req["type"] != tc.requestType {
				t.Fatalf("request: %#v", req)
			}
			delete(req, "id")
			delete(req, "type")
			var payload any = req
			if tc.endpoint != "" {
				if req["endpoint"] != tc.endpoint || req["method"] != tc.method {
					t.Fatalf("proxy: %#v", req)
				}
				if value, ok := req["timeout"]; !ok || value != nil {
					t.Fatalf("proxy timeout: %#v", req)
				}
				payload = req["data"]
			}
			if !reflect.DeepEqual(payload, tc.expected) {
				t.Fatalf("payload %#v want %#v", payload, tc.expected)
			}
			if !strings.Contains(out, `"accepted": true`) {
				t.Fatal(out)
			}
		})
	}
}

func TestManagementInputAndPreviews(t *testing.T) {
	for _, tc := range []struct {
		name, data, file, arg string
		wantError             bool
	}{
		{name: "false", data: `{"value":false}`}, {name: "yaml", data: "value: 0"},
		{name: "null object", data: `null`, wantError: true}, {name: "array", data: `[]`, wantError: true},
		{name: "reserved id", data: `{"id":3}`, wantError: true}, {name: "reserved type", data: `{"type":"other"}`, wantError: true},
		{name: "conflict", data: `{}`, file: "-", wantError: true}, {name: "missing", wantError: true},
		{name: "path traversal", data: `{}`, arg: "../secret", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, connections := appTestServer(t, nil, nil)
			spec := managementSpec{Use: "update <key>", Transport: "ws", Endpoint: "resource/{key}", Args: []string{"key"}, Input: true, RequireInput: true, Effect: "write"}
			cmd := newManagementCommand(spec)
			arg := tc.arg
			if arg == "" {
				arg = "resource"
			}
			args := []string{arg, "--plan"}
			if tc.data != "" {
				args = append(args, "--data", tc.data)
			}
			if tc.file != "" {
				args = append(args, "--file", tc.file)
			}
			out, err := executeAppTest(t, cmd, args...)
			if (err != nil) != tc.wantError {
				t.Fatalf("output=%s error=%v", out, err)
			}
			if connections.Load() != 0 {
				t.Fatal("preview contacted server")
			}
		})
	}
	t.Run("secret redaction", func(t *testing.T) {
		cmd := newManagementCommand(managementSpec{Use: "create", Transport: "ws", Endpoint: "user/create", Input: true, Effect: "write"})
		out, err := executeAppTest(t, cmd, "--data", `{"password":"sensitive-marker"}`, "--plan")
		if err != nil || strings.Contains(out, "sensitive-marker") {
			t.Fatalf("%s %v", out, err)
		}
	})
	t.Run("confirmation", func(t *testing.T) {
		_, connections := appTestServer(t, nil, nil)
		cmd := newManagementCommand(managementSpec{Use: "delete", Transport: "ws", Endpoint: "user/delete", Effect: "destructive"})
		_, err := executeAppTest(t, cmd)
		if err == nil || connections.Load() != 0 {
			t.Fatalf("error=%v connections=%d", err, connections.Load())
		}
	})
}

func TestManagementFlows(t *testing.T) {
	for _, tc := range []struct{ command, args, body, endpoint string }{
		{"integration flow start", "demo", `{"handler":"demo"}`, "config/config_entries/flow"},
		{"integration options start", "entry", `{"handler":"entry"}`, "config/config_entries/options/flow"},
		{"integration subentry flow start", "entry child", `{"handler":["entry","child"]}`, "config/config_entries/subentries/flow"},
		{"repairs flow start", "domain issue", `{"handler":"domain","issue_id":"issue"}`, "repairs/issues/fix"},
	} {
		t.Run(tc.command, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/"+tc.endpoint || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer token" {
					t.Errorf("unexpected %s %s", r.Method, r.URL)
				}
				if strings.HasPrefix(tc.endpoint, "config/") && r.Header.Get("HA-Frontend-Base") == "" {
					t.Error("OAuth frontend header missing")
				}
				var body, expected any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if err := json.Unmarshal([]byte(tc.body), &expected); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(body, expected) {
					t.Errorf("body %#v want %#v", body, expected)
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := fmt.Fprint(w, `{"type":"form","flow_id":"flow","data_schema":[],"errors":{"base":"example"}}`); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			t.Setenv("HAB_URL", server.URL)
			t.Setenv("HAB_TOKEN", "token")
			previous := getAuthManager()
			cachedAuthMgr = auth.NewManager(t.TempDir())
			defer func() { cachedAuthMgr = previous }()
			out, err := executeAppTest(t, managementTestCommand(t, tc.command), strings.Fields(tc.args)...)
			if err != nil || !strings.Contains(out, `"flow_id": "flow"`) || !strings.Contains(out, `"errors"`) {
				t.Fatalf("%s %v", out, err)
			}
		})
	}
}

func TestManagementTimeoutOutcome(t *testing.T) {
	for _, cause := range []error{io.EOF, io.ErrUnexpectedEOF, io.ErrClosedPipe} {
		_, message, details := classifyError(managementMutationError(cause))
		if !strings.Contains(message, "outcome is unknown") || details["retryable"] != false {
			t.Fatalf("%v: %s %#v", cause, message, details)
		}
	}
	err := managementMutationError(context.DeadlineExceeded)
	code, message, details := classifyError(err)
	if code != client.ErrCodeTimeout || !strings.Contains(message, "outcome is unknown") || details["retryable"] != false {
		t.Fatalf("%s %s %#v", code, message, details)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			t.Error(err)
		}
		<-r.Context().Done()
	}))
	defer server.Close()
	t.Setenv("HAB_URL", server.URL)
	t.Setenv("HAB_TOKEN", "token")
	previous := getAuthManager()
	cachedAuthMgr = auth.NewManager(t.TempDir())
	defer func() { cachedAuthMgr = previous }()
	requestCommand := &cobra.Command{}
	requestCommand.SetContext(context.Background())
	_, err = executeManagementRequest(requestCommand, managementSpec{Transport: "rest", Method: "POST", Effect: "write"}, "test", map[string]any{}, 20*time.Millisecond)
	code, message, details = classifyError(err)
	if code != client.ErrCodeTimeout || !strings.Contains(message, "outcome is unknown") || details["retryable"] != false {
		t.Fatalf("%s %s %#v", code, message, details)
	}
}

func TestTransferArguments(t *testing.T) {
	f := &transferFlags{timeout: time.Minute, output: "out", agents: []string{"local", "cloud"}}
	if _, _, _, err := prepareTransfer("backup", false, "backup", f); err == nil {
		t.Fatal("multiple download agents accepted")
	}
	endpoint, query, _, err := prepareTransfer("backup", true, "archive.tar", f)
	if err != nil || endpoint != "backup/upload" || len(query["agent_id"]) != 2 {
		t.Fatalf("%s %v %v", endpoint, query, err)
	}
	file := t.TempDir() + "/password"
	if err := os.WriteFile(file, []byte(" secret \n"), 0600); err != nil {
		t.Fatal(err)
	}
	f.agents = []string{"local"}
	f.passwordFile = file
	_, query, _, err = prepareTransfer("backup", false, "backup", f)
	if err != nil || query.Get("password") != " secret " {
		t.Fatalf("password handling failed: %v", err)
	}
	f.device = "device"
	endpoint, _, _, err = prepareTransfer("diagnostics", false, "entry", f)
	if err != nil || endpoint != "diagnostics/config_entry/entry/device/device" {
		t.Fatalf("%s %v", endpoint, err)
	}
}

// impactTestServer answers the read-only commands an impact plan may send and
// fails the test on any other command, so a plan that writes is caught.
func impactTestServer(t *testing.T, entries []map[string]any) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		var message map[string]any
		if conn.WriteJSON(map[string]any{"type": "auth_required"}) != nil || conn.ReadJSON(&message) != nil || conn.WriteJSON(map[string]any{"type": "auth_ok"}) != nil {
			return
		}
		for {
			message = nil
			if conn.ReadJSON(&message) != nil {
				return
			}
			kind, _ := message["type"].(string)
			if entries[0]["domain"] == "slow" {
				time.Sleep(400 * time.Millisecond)
			}
			var result any
			switch kind {
			case "config_entries/get":
				list := []any{}
				for _, entry := range entries {
					if domain, ok := message["domain"]; !ok || entry["domain"] == domain {
						list = append(list, entry)
					}
				}
				result = list
			case "config_entries/get_single":
				for _, entry := range entries {
					if entry["entry_id"] == message["entry_id"] {
						result = map[string]any{"config_entry": entry}
					}
				}
			case "manifest/list":
				result = []any{
					map[string]any{"domain": "mqtt", "dependencies": []any{"http"}},
					map[string]any{"domain": "frigate", "dependencies": []any{"mqtt", "http"}},
					map[string]any{"domain": "tasmota", "dependencies": []any{"mqtt"}},
					map[string]any{"domain": "zha", "after_dependencies": []any{"mqtt"}},
				}
			case "config/device_registry/list":
				result = []any{
					map[string]any{"id": "dev1", "config_entries": []any{entries[0]["entry_id"]}},
					map[string]any{"id": "shared", "config_entries": []any{entries[0]["entry_id"], "other"}},
					map[string]any{"id": "foreign", "config_entries": []any{"other"}},
				}
			case "config/entity_registry/list":
				result = []any{
					map[string]any{"entity_id": "sensor.b", "config_entry_id": entries[0]["entry_id"], "device_id": "shared"},
					map[string]any{"entity_id": "sensor.a", "config_entry_id": entries[0]["entry_id"], "disabled_by": "user"},
					map[string]any{"entity_id": "light.other", "config_entry_id": "other", "device_id": "dev1"},
				}
			default:
				t.Errorf("unexpected command %s", kind)
			}
			if conn.WriteJSON(map[string]any{"id": message["id"], "type": "result", "success": true, "result": result}) != nil {
				return
			}
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("HAB_URL", server.URL)
	t.Setenv("HAB_TOKEN", "token")
	previous := getAuthManager()
	cachedAuthMgr = auth.NewManager(t.TempDir())
	t.Cleanup(func() { cachedAuthMgr = previous })
}

func TestIntegrationDeletePlanImpact(t *testing.T) {
	const mqtt1, mqtt2 = "0123456789abcdef0123456789abcdef", "fedcba9876543210fedcba9876543210"
	plan := func(t *testing.T, arg string) map[string]any {
		t.Helper()
		out, err := executeAppTest(t, managementTestCommand(t, "integration delete"), arg, "--plan")
		if err != nil {
			t.Fatalf("%s %v", out, err)
		}
		var envelope struct{ Data map[string]any }
		if err := json.Unmarshal([]byte(out), &envelope); err != nil {
			t.Fatal(out)
		}
		return envelope.Data
	}
	t.Run("last entry lists dependents", func(t *testing.T) {
		impactTestServer(t, []map[string]any{{"entry_id": mqtt1, "domain": "mqtt"}, {"entry_id": "other", "domain": "hue"}})
		data := plan(t, mqtt1[:8])
		impact := data["impact"].(map[string]any)
		if impact["status"] != "determined" || impact["entry_id"] != mqtt1 || !reflect.DeepEqual(impact["dependent_integrations"], []any{"frigate", "tasmota"}) || !reflect.DeepEqual(impact["entities"], []any{"sensor.a", "sensor.b"}) || !reflect.DeepEqual(impact["devices"], []any{"dev1"}) {
			t.Fatalf("%#v", impact)
		}
		if risks, _ := data["risks"].([]any); len(risks) != 1 || !strings.Contains(risks[0].(string), "frigate, tasmota") {
			t.Fatalf("%#v", data["risks"])
		}
	})
	t.Run("not the last entry", func(t *testing.T) {
		impactTestServer(t, []map[string]any{{"entry_id": mqtt1, "domain": "mqtt"}, {"entry_id": mqtt2, "domain": "mqtt"}})
		impact := plan(t, mqtt1)["impact"].(map[string]any)
		if impact["other_domain_entries"] != float64(1) || len(impact["dependent_integrations"].([]any)) != 0 {
			t.Fatalf("%#v", impact)
		}
	})
	t.Run("ignored and disabled entries do not count", func(t *testing.T) {
		impactTestServer(t, []map[string]any{{"entry_id": mqtt1, "domain": "mqtt"}, {"entry_id": mqtt2, "domain": "mqtt", "disabled_by": "user"}, {"entry_id": "ignored", "domain": "mqtt", "source": "ignore"}})
		impact := plan(t, mqtt1)["impact"].(map[string]any)
		if impact["other_domain_entries"] != float64(0) || len(impact["dependent_integrations"].([]any)) != 2 {
			t.Fatalf("%#v", impact)
		}
	})
	t.Run("disabled entry breaks nothing", func(t *testing.T) {
		impactTestServer(t, []map[string]any{{"entry_id": mqtt1, "domain": "mqtt", "disabled_by": "user"}})
		if deps := plan(t, mqtt1)["impact"].(map[string]any)["dependent_integrations"].([]any); len(deps) != 0 {
			t.Fatalf("%#v", deps)
		}
	})
	t.Run("slow server is bounded by timeout", func(t *testing.T) {
		impactTestServer(t, []map[string]any{{"entry_id": mqtt1, "domain": "slow"}})
		start := time.Now()
		out, err := executeAppTest(t, managementTestCommand(t, "integration delete"), mqtt1, "--plan", "--timeout", "1s")
		if err != nil || !strings.Contains(out, `"not_determined"`) || time.Since(start) > 1500*time.Millisecond {
			t.Fatalf("%s %v %s", out, err, time.Since(start))
		}
	})
	t.Run("no dependents", func(t *testing.T) {
		impactTestServer(t, []map[string]any{{"entry_id": mqtt1, "domain": "frigate"}})
		data := plan(t, mqtt1)
		if deps := data["impact"].(map[string]any)["dependent_integrations"].([]any); len(deps) != 0 || data["risks"] != nil {
			t.Fatalf("%#v", data)
		}
	})
	t.Run("server unreachable", func(t *testing.T) {
		t.Setenv("HAB_URL", "http://127.0.0.1:1")
		t.Setenv("HAB_TOKEN", "token")
		previous := getAuthManager()
		cachedAuthMgr = auth.NewManager(t.TempDir())
		defer func() { cachedAuthMgr = previous }()
		data := plan(t, mqtt1)
		impact := data["impact"].(map[string]any)
		if data["mode"] != "plan" || impact["status"] != "not_determined" || impact["reason"] == "" {
			t.Fatalf("%#v", data)
		}
	})
}

func TestIntegrationGetDependents(t *testing.T) {
	impactTestServer(t, []map[string]any{{"entry_id": "entry", "domain": "mqtt"}})
	out, err := executeAppTest(t, managementTestCommand(t, "integration get"), "entry")
	if err != nil || !strings.Contains(out, `"dependent_integrations": [`) || !strings.Contains(out, `"frigate"`) || strings.Contains(out, `"zha"`) {
		t.Fatalf("%s %v", out, err)
	}
}

func TestPlanImpactText(t *testing.T) {
	got := planImpactText(MutationPlan{Risks: []string{"frigate will break"}, Impact: map[string]any{"status": "determined", "domain": "mqtt", "other_domain_entries": 0, "dependent_integrations": []string{"frigate"}, "devices": []string{}, "entities": []string{"sensor.a", "sensor.b"}}})
	want := "\nRisk: frigate will break\nImpact: determined\n  dependent_integrations (1): frigate\n  devices (0): \n  domain: mqtt\n  entities (2): sensor.a, sensor.b\n  other_domain_entries: 0"
	if got != want {
		t.Fatalf("%q", got)
	}
	got = planImpactText(MutationPlan{Impact: map[string]any{"status": "not_determined", "reason": "connection refused"}})
	if got != "\nImpact: not_determined\n  reason: connection refused" {
		t.Fatalf("%q", got)
	}
}

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

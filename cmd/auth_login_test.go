package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/config"
	"github.com/spf13/viper"
)

func useLoginManager(t *testing.T, m *auth.Manager) {
	t.Helper()
	previous := getAuthManager()
	cachedAuthMgr = m
	t.Cleanup(func() { cachedAuthMgr = previous })
	oldToken, oldURL, oldAccess := loginToken, loginURL, loginAccessToken
	t.Cleanup(func() { loginToken, loginURL, loginAccessToken = oldToken, oldURL, oldAccess })
	viper.Set("text", false)
	t.Cleanup(func() { viper.Set("text", nil) })
}

func TestAuthLoginRejectsUnresolvableContextBeforePrompting(t *testing.T) {
	for _, name := range []string{"HAB_URL", "HAB_TOKEN", "HAB_REFRESH_TOKEN", config.ContextEnv} {
		t.Setenv(name, "")
	}
	badCurrent := t.TempDir()
	if err := os.WriteFile(filepath.Join(badCurrent, config.CurrentContextFile), []byte("a/b\n"), 0600); err != nil {
		t.Fatal(err)
	}

	for name, m := range map[string]*auth.Manager{
		"flag":            auth.NewContextManager(t.TempDir(), "a/b"),
		"current-context": auth.NewContextManager(badCurrent, ""),
	} {
		t.Run(name, func(t *testing.T) {
			useLoginManager(t, m)
			// No URL or token: any prompt, discovery, or network step would block or fail differently.
			loginToken, loginURL, loginAccessToken = true, "", ""
			err := runAuthLogin(authLoginCmd, nil)
			if err == nil || !strings.Contains(err.Error(), "invalid context name") {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestAuthLoginWarnsWhenEnvCredentialsShadowContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"location_name":"Brother","version":"2026.11.0"}`))
	}))
	defer server.Close()
	t.Setenv("HAB_URL", "http://env-instance")
	t.Setenv("HAB_TOKEN", "env-token")
	t.Setenv("HAB_REFRESH_TOKEN", "")
	t.Setenv(config.ContextEnv, "")

	root := t.TempDir()
	useLoginManager(t, auth.NewContextManager(root, "brother"))
	resetExecutionMetadata()
	loginToken, loginURL, loginAccessToken = true, server.URL, "brother-token"
	if err := runAuthLogin(authLoginCmd, nil); err != nil {
		t.Fatal(err)
	}

	warnings, _ := getExecutionMetadata()["warnings"].([]string)
	if len(warnings) != 1 || !strings.Contains(warnings[0], `use context "brother"`) {
		t.Fatalf("warnings = %v", warnings)
	}
	stored, err := auth.LoadStoredCredentials(config.GetContextDir(root, "brother"))
	if err != nil || stored == nil || stored.URL != server.URL {
		t.Fatalf("stored = %+v, %v", stored, err)
	}
	// The documented consequence: the same environment cannot use the context.
	if _, err := auth.NewContextManager(root, "brother").GetCredentials(); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("err = %v", err)
	}
}

func TestAuthContextListFlagsUnreadableContext(t *testing.T) {
	for _, name := range []string{"HAB_URL", "HAB_TOKEN", "HAB_REFRESH_TOKEN", config.ContextEnv} {
		t.Setenv(name, "")
	}
	root := t.TempDir()
	viper.Set("config", root)
	t.Cleanup(func() { viper.Set("config", nil) })
	useLoginManager(t, auth.NewContextManager(root, ""))
	if err := auth.SaveCredentials(&auth.Credentials{URL: "http://home:8123", AccessToken: "home"}, root); err != nil {
		t.Fatal(err)
	}
	broken := config.GetContextDir(root, "brother")
	if err := os.MkdirAll(broken, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config.GetCredentialsPath(broken), []byte(`{"url":"http://brother:8123","data":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}`), 0600); err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = file
	err = runAuthContextList(authContextListCmd, nil)
	os.Stdout = stdout
	if closeErr := file.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(file.Name())
	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope.Data) != 2 {
		t.Fatalf("output = %s, %v", raw, err)
	}
	home, brother := envelope.Data[0], envelope.Data[1]
	if brother["name"] != "brother" || brother["unreadable"] != true || brother["url"] != "http://brother:8123" {
		t.Fatalf("brother = %#v", brother)
	}
	if home["name"] != "default" || home["unreadable"] != nil || home["url"] != "http://home:8123" {
		t.Fatalf("default = %#v", home)
	}
}

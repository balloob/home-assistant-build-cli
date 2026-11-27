package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/home-assistant/hab/config"
)

// clearAuthEnv isolates a test from credentials in the caller's environment.
func clearAuthEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{"HAB_URL", "HAB_TOKEN", "HAB_REFRESH_TOKEN", config.ContextEnv, SupervisorTokenEnv} {
		t.Setenv(name, "")
	}
}

func saveContext(t *testing.T, root, name string, creds *Credentials) {
	t.Helper()
	if err := SaveCredentials(creds, config.GetContextDir(root, name)); err != nil {
		t.Fatalf("save %s: %v", name, err)
	}
}

func TestLegacyCredentialsAreDefaultContext(t *testing.T) {
	clearAuthEnv(t)
	root := t.TempDir()
	// A pre-contexts install wrote credentials.json directly into the config dir.
	if err := SaveCredentials(&Credentials{URL: "http://home:8123", AccessToken: "home"}, root); err != nil {
		t.Fatal(err)
	}

	m := NewContextManager(root, "")
	creds, err := m.GetCredentials()
	if err != nil {
		t.Fatal(err)
	}
	if m.Context != config.DefaultContext || creds.URL != "http://home:8123" || creds.Source != "credentials_file" {
		t.Fatalf("context=%q creds=%+v", m.Context, creds)
	}
	names, err := config.ListContexts(root)
	if err != nil || len(names) != 1 || names[0] != config.DefaultContext {
		t.Fatalf("ListContexts = %v, %v", names, err)
	}
	if _, err := os.Stat(filepath.Join(root, config.CurrentContextFile)); !os.IsNotExist(err) {
		t.Fatal("reading a legacy install must not write a current-context file")
	}
}

func TestContextSelectionPrecedence(t *testing.T) {
	clearAuthEnv(t)
	root := t.TempDir()
	saveContext(t, root, config.DefaultContext, &Credentials{URL: "http://home", AccessToken: "home"})
	saveContext(t, root, "brother", &Credentials{URL: "http://brother", AccessToken: "brother"})
	saveContext(t, root, "office", &Credentials{URL: "http://office", AccessToken: "office"})

	urlFor := func(flag string) string {
		t.Helper()
		creds, err := NewContextManager(root, flag).GetCredentials()
		if err != nil {
			t.Fatal(err)
		}
		return creds.URL
	}

	if got := urlFor(""); got != "http://home" {
		t.Fatalf("no selection = %s", got)
	}
	if err := config.SaveCurrentContext(root, "office"); err != nil {
		t.Fatal(err)
	}
	if got := urlFor(""); got != "http://office" {
		t.Fatalf("current = %s", got)
	}
	t.Setenv(config.ContextEnv, "brother")
	if got := urlFor(""); got != "http://brother" {
		t.Fatalf("env over current = %s", got)
	}
	if got := urlFor("default"); got != "http://home" {
		t.Fatalf("flag over env = %s", got)
	}
}

func TestUnknownContextNeverFallsBack(t *testing.T) {
	clearAuthEnv(t)
	root := t.TempDir()
	saveContext(t, root, config.DefaultContext, &Credentials{URL: "http://home", AccessToken: "home"})
	saveContext(t, root, "brother", &Credentials{URL: "http://brother", AccessToken: "brother"})
	t.Setenv(SupervisorTokenEnv, "supervisor-token")

	for name, m := range map[string]*Manager{
		"flag": NewContextManager(root, "typo"),
		"current": func() *Manager {
			if err := config.SaveCurrentContext(root, "gone"); err != nil {
				t.Fatal(err)
			}
			return NewContextManager(root, "")
		}(),
	} {
		creds, err := m.GetCredentials()
		if creds != nil || !errors.Is(err, ErrNotAuthenticated) {
			t.Fatalf("%s: creds=%+v err=%v", name, creds, err)
		}
		if !strings.Contains(err.Error(), "available contexts: default, brother") {
			t.Fatalf("%s: error does not list contexts: %v", name, err)
		}
	}

	if _, err := NewContextManager(root, "../etc").GetCredentials(); err == nil || !strings.Contains(err.Error(), "invalid context name") {
		t.Fatalf("path-like name accepted: %v", err)
	}
}

func TestExplicitContextConflictsWithEnvCredentials(t *testing.T) {
	clearAuthEnv(t)
	root := t.TempDir()
	saveContext(t, root, "brother", &Credentials{URL: "http://brother", AccessToken: "brother"})
	if err := config.SaveCurrentContext(root, "brother"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HAB_URL", "http://env")
	t.Setenv("HAB_TOKEN", "env")

	// Without an explicit selection, env credentials win as they always have.
	creds, err := NewContextManager(root, "").GetCredentials()
	if err != nil || creds.URL != "http://env" {
		t.Fatalf("implicit: creds=%+v err=%v", creds, err)
	}

	if _, err := NewContextManager(root, "brother").GetCredentials(); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("flag: err=%v", err)
	}
	t.Setenv(config.ContextEnv, "brother")
	if _, err := NewContextManager(root, "").GetCredentials(); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("env: err=%v", err)
	}
}

func TestSupervisorFallbackOnlyForDefaultContext(t *testing.T) {
	clearAuthEnv(t)
	root := t.TempDir()
	t.Setenv(SupervisorTokenEnv, "supervisor-token")

	creds, err := NewContextManager(root, "").GetCredentials()
	if err != nil || creds.Source != "supervisor" || creds.URL != SupervisorURL {
		t.Fatalf("default: creds=%+v err=%v", creds, err)
	}
	if _, err := NewContextManager(root, "brother").GetCredentials(); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("named: err=%v", err)
	}
}

func TestTokenRefreshWritesToSelectedContext(t *testing.T) {
	clearAuthEnv(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"fresh","expires_in":1800,"token_type":"Bearer"}`))
	}))
	defer server.Close()

	root := t.TempDir()
	home := &Credentials{URL: "http://home", AccessToken: "home"}
	saveContext(t, root, config.DefaultContext, home)
	saveContext(t, root, "brother", &Credentials{URL: server.URL, ClientID: "id", AccessToken: "stale", RefreshToken: "refresh", TokenExpiry: 1})

	creds, err := NewContextManager(root, "brother").GetCredentials()
	if err != nil || creds.AccessToken != "fresh" {
		t.Fatalf("refresh: creds=%+v err=%v", creds, err)
	}

	stored, err := LoadStoredCredentials(config.GetContextDir(root, "brother"))
	if err != nil || stored.AccessToken != "fresh" || stored.RefreshToken != "refresh" {
		t.Fatalf("brother not updated: %+v %v", stored, err)
	}
	stored, err = LoadStoredCredentials(root)
	if err != nil || *stored != (Credentials{URL: home.URL, AccessToken: home.AccessToken, Source: "credentials_file"}) {
		t.Fatalf("default context changed: %+v %v", stored, err)
	}
}

func TestEnvRefreshTokenIsNotPersisted(t *testing.T) {
	clearAuthEnv(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"fresh","expires_in":1800,"token_type":"Bearer"}`))
	}))
	defer server.Close()
	root := t.TempDir()
	saveContext(t, root, "brother", &Credentials{URL: "http://brother", AccessToken: "brother"})
	if err := config.SaveCurrentContext(root, "brother"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HAB_URL", server.URL)
	t.Setenv("HAB_REFRESH_TOKEN", "env-refresh")

	if _, err := NewContextManager(root, "").GetCredentials(); err != nil {
		t.Fatal(err)
	}
	stored, err := LoadStoredCredentials(config.GetContextDir(root, "brother"))
	if err != nil || stored.URL != "http://brother" || stored.AccessToken != "brother" {
		t.Fatalf("env refresh overwrote the brother context: %+v %v", stored, err)
	}
}

func TestLogoutOnlyAffectsSelectedContext(t *testing.T) {
	clearAuthEnv(t)
	root := t.TempDir()
	saveContext(t, root, config.DefaultContext, &Credentials{URL: "http://home", AccessToken: "home"})
	saveContext(t, root, "brother", &Credentials{URL: "http://brother", AccessToken: "brother"})
	brotherDir := config.GetContextDir(root, "brother")
	if err := config.SaveESPHomeContext(brotherDir, &config.ESPHomeContext{Configuration: "garage.yaml"}); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveESPHomeContext(root, &config.ESPHomeContext{Configuration: "kitchen.yaml"}); err != nil {
		t.Fatal(err)
	}

	if err := NewContextManager(root, "brother").Logout(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(brotherDir); !os.IsNotExist(err) {
		t.Fatalf("brother context still exists: %v", err)
	}
	creds, err := NewContextManager(root, "").GetCredentials()
	if err != nil || creds.URL != "http://home" {
		t.Fatalf("default context lost: %+v %v", creds, err)
	}

	saveContext(t, root, "brother", &Credentials{URL: "http://brother", AccessToken: "brother"})
	if err := NewContextManager(root, config.DefaultContext).Logout(); err != nil {
		t.Fatal(err)
	}
	if creds, _ := LoadStoredCredentials(root); creds != nil {
		t.Fatal("default credentials not removed")
	}
	if ctx, _ := config.LoadESPHomeContext(root); ctx == nil {
		t.Fatal("default logout removed more than credentials")
	}
	creds, err = NewContextManager(root, "brother").GetCredentials()
	if err != nil || creds.URL != "http://brother" {
		t.Fatalf("brother context lost: %+v %v", creds, err)
	}
}

func TestSaveWritesToSelectedContext(t *testing.T) {
	clearAuthEnv(t)
	root := t.TempDir()
	if err := NewContextManager(root, "brother").Save(&Credentials{URL: "http://brother", AccessToken: "brother"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, config.CredentialsFile)); !os.IsNotExist(err) {
		t.Fatal("login to a named context wrote default credentials")
	}
	if current, _ := config.LoadCurrentContext(root); current != config.DefaultContext {
		t.Fatalf("login changed the current context to %q", current)
	}
	status := NewContextManager(root, "brother").GetAuthStatus()
	if status["authenticated"] != true || status["context"] != "brother" || status["url"] != "http://brother" {
		t.Fatalf("status = %v", status)
	}
}

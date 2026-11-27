package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetContextDir(t *testing.T) {
	if got := GetContextDir("/cfg", DefaultContext); got != "/cfg" {
		t.Fatalf("default = %q", got)
	}
	if got := GetContextDir("/cfg", "brother"); got != filepath.Join("/cfg", ContextsDir, "brother") {
		t.Fatalf("named = %q", got)
	}
}

func TestValidateContextName(t *testing.T) {
	for _, name := range []string{"default", "brother-homeassistant", "a.b_c", "a1"} {
		if err := ValidateContextName(name); err != nil {
			t.Errorf("%q rejected: %v", name, err)
		}
	}
	for _, name := range []string{"", ".", "..", "../x", "a/b", `a\b`, "-x", ".hidden", "with space", "A1", "Default", "DEFAULT", "brother."} {
		if err := ValidateContextName(name); err == nil {
			t.Errorf("%q accepted", name)
		}
	}
}

func TestResolveContextPrecedence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(ContextEnv, "")

	check := func(flag, wantName, wantSource string) {
		t.Helper()
		name, source, err := ResolveContext(dir, flag)
		if err != nil || name != wantName || source != wantSource {
			t.Fatalf("ResolveContext(%q) = %q, %q, %v; want %q, %q", flag, name, source, err, wantName, wantSource)
		}
	}

	check("", DefaultContext, ContextSourceDefault)
	if err := SaveCurrentContext(dir, "office"); err != nil {
		t.Fatal(err)
	}
	check("", "office", ContextSourceCurrent)
	t.Setenv(ContextEnv, "brother")
	check("", "brother", ContextSourceEnv)
	check("home", "home", ContextSourceFlag)

	t.Setenv(ContextEnv, "../escape")
	if _, _, err := ResolveContext(dir, ""); err == nil {
		t.Fatal("invalid HAB_CONTEXT accepted")
	}
}

func TestSaveCurrentContextDefaultRemovesFile(t *testing.T) {
	dir := t.TempDir()
	if err := SaveCurrentContext(dir, "brother"); err != nil {
		t.Fatal(err)
	}
	if got, _ := LoadCurrentContext(dir); got != "brother" {
		t.Fatalf("current = %q", got)
	}
	if err := SaveCurrentContext(dir, DefaultContext); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, CurrentContextFile)); !os.IsNotExist(err) {
		t.Fatalf("current-context file kept: %v", err)
	}
}

func TestListContexts(t *testing.T) {
	dir := t.TempDir()
	if names, err := ListContexts(dir); err != nil || len(names) != 0 {
		t.Fatalf("empty = %v, %v", names, err)
	}
	for _, name := range []string{"zeta", "alpha"} {
		ctxDir := GetContextDir(dir, name)
		if err := os.MkdirAll(ctxDir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(GetCredentialsPath(ctxDir), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	// A directory without credentials (for example only ESPHome state) is not a context.
	if err := os.MkdirAll(GetContextDir(dir, "empty"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(GetCredentialsPath(dir), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	names, err := ListContexts(dir)
	if err != nil || len(names) != 3 || names[0] != DefaultContext || names[1] != "alpha" || names[2] != "zeta" {
		t.Fatalf("ListContexts = %v, %v", names, err)
	}
}

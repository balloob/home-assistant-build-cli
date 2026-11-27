package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/home-assistant/hab/internal/fileutil"
)

const (
	// DefaultContext is the context stored directly in the config directory,
	// where single-instance installs have always kept their credentials.
	DefaultContext = "default"
	// ContextsDir holds one subdirectory per named context.
	ContextsDir = "contexts"
	// CurrentContextFile holds the name of the current context.
	CurrentContextFile = "current-context"
	// ContextEnv overrides the current context for one invocation.
	ContextEnv = "HAB_CONTEXT"
)

// Context sources reported by ResolveContext.
const (
	ContextSourceFlag    = "flag"
	ContextSourceEnv     = "env"
	ContextSourceCurrent = "current"
	ContextSourceDefault = "default"
)

// Lowercase only and no trailing '.', because Windows and macOS file systems
// would otherwise map "Brother" or "brother." onto the "brother" directory.
var contextNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

// ValidateContextName rejects names that are not safe as a single path component.
func ValidateContextName(name string) error {
	if !contextNamePattern.MatchString(name) || strings.HasSuffix(name, ".") {
		return fmt.Errorf("invalid context name %q: use 1-64 lowercase letters, digits, '.', '_' or '-', starting with a letter or digit and not ending with '.'", name)
	}
	return nil
}

// GetContextDir returns the directory holding a context's credentials and
// per-instance state. The default context is the config directory itself.
func GetContextDir(configDir, name string) string {
	if name == "" || name == DefaultContext {
		return GetConfigDir(configDir)
	}
	return filepath.Join(GetConfigDir(configDir), ContextsDir, name)
}

// LoadCurrentContext returns the saved current context, or DefaultContext.
func LoadCurrentContext(configDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(GetConfigDir(configDir), CurrentContextFile))
	if os.IsNotExist(err) {
		return DefaultContext, nil
	}
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(data))
	if name == "" {
		return DefaultContext, nil
	}
	if err := ValidateContextName(name); err != nil {
		return "", fmt.Errorf("%s file: %w", CurrentContextFile, err)
	}
	return name, nil
}

// SaveCurrentContext makes name the current context. Selecting the default
// context removes the file so older hab versions see an unchanged layout.
func SaveCurrentContext(configDir, name string) error {
	if err := ValidateContextName(name); err != nil {
		return err
	}
	path := filepath.Join(GetConfigDir(configDir), CurrentContextFile)
	if name == DefaultContext {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err := EnsureConfigDir(configDir); err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(path, []byte(name+"\n"), 0600)
}

// ResolveContext picks the context for this invocation:
// flag > HAB_CONTEXT > current-context file > default.
func ResolveContext(configDir, flagValue string) (name, source string, err error) {
	name, source = flagValue, ContextSourceFlag
	if name == "" {
		name, source = os.Getenv(ContextEnv), ContextSourceEnv
	}
	if name != "" {
		return name, source, ValidateContextName(name)
	}
	name, err = LoadCurrentContext(configDir)
	if err != nil {
		return "", "", err
	}
	if name == DefaultContext {
		return name, ContextSourceDefault, nil
	}
	return name, ContextSourceCurrent, nil
}

// ListContexts returns the contexts that have stored credentials, sorted,
// with the default context first when present.
func ListContexts(configDir string) ([]string, error) {
	var names []string
	if _, err := os.Stat(GetCredentialsPath(configDir)); err == nil {
		names = append(names, DefaultContext)
	}
	entries, err := os.ReadDir(filepath.Join(GetConfigDir(configDir), ContextsDir))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var named []string
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || name == DefaultContext || ValidateContextName(name) != nil {
			continue
		}
		if _, err := os.Stat(GetCredentialsPath(GetContextDir(configDir, name))); err == nil {
			named = append(named, name)
		}
	}
	sort.Strings(named)
	return append(names, named...), nil
}

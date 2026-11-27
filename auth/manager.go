// Package auth handles authentication for Home Assistant, including OAuth
// flows, long-lived access tokens, Supervisor add-on credentials, encrypted
// credential storage, and mDNS server discovery.
package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/config"
	log "github.com/sirupsen/logrus"
)

// ErrNotAuthenticated is returned when the user is not authenticated
var ErrNotAuthenticated = errors.New("not authenticated")

// Manager handles authentication state and token refresh for one context
type Manager struct {
	ConfigDir     string
	Context       string
	ContextSource string
	contextErr    error
	credentials   *Credentials
}

// NewManager creates an auth manager for the default context
func NewManager(configDir string) *Manager {
	return &Manager{
		ConfigDir:     configDir,
		Context:       config.DefaultContext,
		ContextSource: config.ContextSourceDefault,
	}
}

// NewContextManager creates an auth manager for the context selected by
// contextFlag, HAB_CONTEXT, or the current-context file, in that order.
// A resolution error is returned by every method that needs the context.
func NewContextManager(configDir, contextFlag string) *Manager {
	name, source, err := config.ResolveContext(configDir, contextFlag)
	return &Manager{
		ConfigDir:     configDir,
		Context:       name,
		ContextSource: source,
		contextErr:    err,
	}
}

// ContextDir returns the directory of the selected context.
func (m *Manager) ContextDir() (string, error) {
	if m.contextErr != nil {
		return "", m.contextErr
	}
	return config.GetContextDir(m.ConfigDir, m.Context), nil
}

func (m *Manager) contextIsExplicit() bool {
	return m.ContextSource == config.ContextSourceFlag || m.ContextSource == config.ContextSourceEnv
}

// EnvCredentialsConflict reports whether the context was selected explicitly
// while HAB_URL with HAB_TOKEN or HAB_REFRESH_TOKEN is also set, which makes
// every credential lookup fail.
func (m *Manager) EnvCredentialsConflict() bool {
	return m.contextIsExplicit() && loadEnvCredentials() != nil
}

// loadCredentials loads credentials without printing warnings (internal use)
func (m *Manager) loadCredentials() (*Credentials, error) {
	if m.credentials != nil {
		return m.credentials, nil
	}
	dir, err := m.ContextDir()
	if err != nil {
		return nil, err
	}
	// An explicit context must never be silently replaced by env credentials.
	if m.EnvCredentialsConflict() {
		return nil, fmt.Errorf("conflicting credentials: context %q was selected by --context or %s, but HAB_URL with HAB_TOKEN or HAB_REFRESH_TOKEN is also set; unset one of them", m.Context, config.ContextEnv)
	}

	var creds *Credentials
	if m.Context == config.DefaultContext {
		creds, err = LoadCredentials(dir)
	} else if creds = loadEnvCredentials(); creds == nil {
		creds, err = LoadStoredCredentials(dir)
		if err == nil && creds == nil {
			err = m.contextNotFound()
		}
	}
	if err != nil {
		return nil, err
	}
	m.credentials = creds
	return creds, nil
}

func (m *Manager) contextNotFound() error {
	available := "none"
	if names, err := config.ListContexts(m.ConfigDir); err == nil && len(names) > 0 {
		available = strings.Join(names, ", ")
	}
	return fmt.Errorf("%w: context %q has no stored credentials (available contexts: %s). Run 'hab auth login --context %s' or 'hab auth context use <name>'", ErrNotAuthenticated, m.Context, available, m.Context)
}

// saveRefreshed persists refreshed credentials only to the context file they
// were loaded from; env and Supervisor credentials are never written to disk.
func (m *Manager) saveRefreshed(creds *Credentials) error {
	if creds.Source != "credentials_file" {
		return nil
	}
	dir, err := m.ContextDir()
	if err != nil {
		return err
	}
	return SaveCredentials(creds, dir)
}

// GetCredentials returns the current credentials, printing a warning if not authenticated.
// It automatically refreshes OAuth tokens if they are expired or about to expire.
func (m *Manager) GetCredentials() (*Credentials, error) {
	creds, err := m.loadCredentials()
	if err != nil {
		return nil, err
	}
	if creds == nil {
		return nil, ErrNotAuthenticated
	}

	// Automatically refresh OAuth tokens if needed
	if creds.NeedsRefresh() && creds.RefreshToken != "" {
		newCreds, err := RefreshAccessToken(creds)
		if err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
		newCreds.Source = creds.Source
		m.credentials = newCreds
		if err := m.saveRefreshed(newCreds); err != nil {
			// Log warning but continue with refreshed credentials in memory
			log.WithError(err).Warn("Failed to save refreshed credentials")
		}
		return newCreds, nil
	}

	return creds, nil
}

// IsAuthenticated returns true if authenticated
func (m *Manager) IsAuthenticated() bool {
	creds, err := m.loadCredentials()
	if err != nil || creds == nil {
		return false
	}
	return creds.HasValidToken()
}

// RefreshToken forces an OAuth token refresh.
// Uses loadCredentials (not GetCredentials) to avoid an automatic refresh
// before the explicit one requested by the caller.
func (m *Manager) RefreshToken() error {
	creds, err := m.loadCredentials()
	if err != nil {
		return err
	}
	if creds == nil || creds.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	newCreds, err := RefreshAccessToken(creds)
	if err != nil {
		return err
	}
	newCreds.Source = creds.Source

	m.credentials = newCreds
	return m.saveRefreshed(newCreds)
}

// Save saves new credentials to the selected context
func (m *Manager) Save(creds *Credentials) error {
	dir, err := m.ContextDir()
	if err != nil {
		return err
	}
	m.credentials = creds
	return SaveCredentials(creds, dir)
}

// Logout removes the selected context's stored credentials. A named context
// is removed entirely, including its per-instance state.
func (m *Manager) Logout() error {
	dir, err := m.ContextDir()
	if err != nil {
		return err
	}
	m.credentials = nil
	if m.Context == config.DefaultContext {
		return DeleteCredentials(dir)
	}
	return os.RemoveAll(dir)
}

// GetAuthStatus returns authentication status as a map
func (m *Manager) GetAuthStatus() map[string]interface{} {
	creds, err := m.loadCredentials()
	if creds == nil {
		message := "Not authenticated. Run 'hab auth login' to authenticate."
		if err != nil {
			message = err.Error()
		}
		return map[string]interface{}{
			"authenticated": false,
			"context":       m.Context,
			"message":       message,
		}
	}

	authType := "token"
	if IsSupervisorEnvironment() {
		authType = "supervisor"
	} else if creds.IsOAuth() {
		authType = "oauth"
	}

	return map[string]interface{}{
		"authenticated":     true,
		"context":           m.Context,
		"url":               creds.URL,
		"auth_type":         authType,
		"credential_source": creds.Source,
		"token_expiry":      creds.TokenExpiry,
	}
}

// GetRestClient returns a configured REST client.
// Token refresh is handled automatically by GetCredentials().
func (m *Manager) GetRestClient() (*client.RestClient, error) {
	creds, err := m.GetCredentials()
	if err != nil {
		return nil, err
	}

	return client.NewRestClient(creds.URL, creds.AccessToken), nil
}

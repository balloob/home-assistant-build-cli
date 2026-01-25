package auth

import (
	"errors"
	"fmt"

	"github.com/home-assistant/hab/client"
)

// ErrNotAuthenticated is returned when the user is not authenticated
var ErrNotAuthenticated = errors.New("not authenticated")

// Manager handles authentication state and token refresh
type Manager struct {
	ConfigDir   string
	credentials *Credentials
}

// NewManager creates a new auth manager
func NewManager(configDir string) *Manager {
	return &Manager{
		ConfigDir: configDir,
	}
}

// loadCredentials loads credentials without printing warnings (internal use)
func (m *Manager) loadCredentials() (*Credentials, error) {
	if m.credentials == nil {
		creds, err := LoadCredentials(m.ConfigDir)
		if err != nil {
			return nil, err
		}
		m.credentials = creds
	}
	return m.credentials, nil
}

// GetCredentials returns the current credentials, printing a warning if not authenticated.
// It automatically refreshes OAuth tokens if they are expired or about to expire.
func (m *Manager) GetCredentials() (*Credentials, error) {
	creds, err := m.loadCredentials()
	if err != nil {
		return nil, err
	}
	if creds == nil {
		client.PrintWarning("Not authenticated. Run 'hab auth login' to authenticate.")
		return nil, ErrNotAuthenticated
	}

	// Automatically refresh OAuth tokens if needed
	if creds.NeedsRefresh() && creds.RefreshToken != "" {
		newCreds, err := RefreshAccessToken(creds)
		if err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
		m.credentials = newCreds
		if err := SaveCredentials(newCreds, m.ConfigDir); err != nil {
			// Log warning but continue with refreshed credentials in memory
			client.PrintWarning("Failed to save refreshed credentials: " + err.Error())
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

// GetURL returns the Home Assistant URL
func (m *Manager) GetURL() string {
	creds, _ := m.loadCredentials()
	if creds == nil {
		return ""
	}
	return creds.URL
}

// GetToken returns the access token
func (m *Manager) GetToken() string {
	creds, _ := m.loadCredentials()
	if creds == nil {
		return ""
	}
	return creds.AccessToken
}

// NeedsRefresh returns true if the token needs to be refreshed
func (m *Manager) NeedsRefresh() bool {
	creds, _ := m.loadCredentials()
	if creds == nil {
		return false
	}
	return creds.NeedsRefresh()
}

// RefreshToken refreshes an OAuth token
func (m *Manager) RefreshToken() error {
	creds, err := m.GetCredentials()
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

	m.credentials = newCreds
	return SaveCredentials(newCreds, m.ConfigDir)
}

// Save saves new credentials
func (m *Manager) Save(creds *Credentials) error {
	m.credentials = creds
	return SaveCredentials(creds, m.ConfigDir)
}

// Logout removes stored credentials
func (m *Manager) Logout() bool {
	m.credentials = nil
	return DeleteCredentials(m.ConfigDir)
}

// GetAuthStatus returns authentication status as a map
func (m *Manager) GetAuthStatus() map[string]interface{} {
	creds, _ := m.loadCredentials()
	if creds == nil {
		return map[string]interface{}{
			"authenticated": false,
			"message":       "Not authenticated. Run 'hab auth login' to authenticate.",
		}
	}

	authType := "token"
	if creds.IsOAuth() {
		authType = "oauth"
	}

	return map[string]interface{}{
		"authenticated": true,
		"url":           creds.URL,
		"auth_type":     authType,
		"token_expiry":  creds.TokenExpiry,
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

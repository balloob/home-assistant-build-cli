package auth

import (
	"encoding/json"
	"os"
	"time"

	"github.com/home-assistant/hab/config"
	"github.com/home-assistant/hab/internal/fileutil"
)

const (
	// tokenRefreshWindow is the number of seconds before expiry at which
	// a token is considered due for refresh (5 minutes).
	tokenRefreshWindow = 300

	// defaultTokenExpiry is the assumed token lifetime in seconds when the
	// server does not provide an expires_in value (30 minutes).
	defaultTokenExpiry = 1800
)

// Credentials stores authentication information
type Credentials struct {
	URL          string  `json:"url"`
	ClientID     string  `json:"client_id,omitempty"`
	AccessToken  string  `json:"access_token,omitempty"`
	RefreshToken string  `json:"refresh_token,omitempty"`
	TokenExpiry  float64 `json:"token_expiry,omitempty"`
	Source       string  `json:"-"`
}

// IsOAuth returns true if using OAuth authentication
func (c *Credentials) IsOAuth() bool {
	return c.RefreshToken != ""
}

// HasValidToken returns true if there is an access token
func (c *Credentials) HasValidToken() bool {
	return c.AccessToken != ""
}

// IsExpired returns true if the token is expired
func (c *Credentials) IsExpired() bool {
	if !c.IsOAuth() {
		return false // Long-lived tokens don't expire
	}
	if c.TokenExpiry == 0 {
		return true // No expiry info, assume expired
	}
	return time.Now().Unix() >= int64(c.TokenExpiry)
}

// NeedsRefresh returns true if the token needs to be refreshed
func (c *Credentials) NeedsRefresh() bool {
	if !c.IsOAuth() {
		return false
	}
	if c.TokenExpiry == 0 {
		return true
	}
	// Refresh if within tokenRefreshWindow of expiry
	return float64(time.Now().Unix()) >= (c.TokenExpiry - tokenRefreshWindow)
}

// LoadCredentials loads credentials for the default context: environment
// variables first, then the encrypted credentials file, then the Supervisor.
func LoadCredentials(configDir string) (*Credentials, error) {
	if creds := loadEnvCredentials(); creds != nil {
		return creds, nil
	}

	creds, err := LoadStoredCredentials(configDir)
	if err != nil || creds != nil {
		return creds, err
	}

	// Fall back to Home Assistant Supervisor environment (add-on/app)
	if token := GetSupervisorToken(); token != "" {
		return &Credentials{
			URL:         SupervisorURL,
			AccessToken: token,
			Source:      "supervisor",
		}, nil
	}

	return nil, nil
}

// loadEnvCredentials returns credentials from HAB_URL with HAB_TOKEN or
// HAB_REFRESH_TOKEN, or nil when they are not set.
func loadEnvCredentials() *Credentials {
	envURL := os.Getenv("HAB_URL")
	if envURL == "" {
		return nil
	}
	if envToken := os.Getenv("HAB_TOKEN"); envToken != "" {
		return &Credentials{
			URL:         envURL,
			AccessToken: envToken,
			Source:      "env_token",
		}
	}
	if envRefresh := os.Getenv("HAB_REFRESH_TOKEN"); envRefresh != "" {
		return &Credentials{
			URL:          envURL,
			RefreshToken: envRefresh,
			Source:       "env_refresh_token",
		}
	}
	return nil
}

// LoadStoredCredentials reads the encrypted credentials file in dir, ignoring
// environment variables and the Supervisor. It returns nil when the file does
// not exist.
func LoadStoredCredentials(dir string) (*Credentials, error) {
	data, err := os.ReadFile(config.GetCredentialsPath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	decrypted, err := decrypt(data, deriveKey())
	if err != nil {
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(decrypted, &creds); err != nil {
		return nil, err
	}

	creds.Source = "credentials_file"
	return &creds, nil
}

// SaveCredentials saves credentials to encrypted storage
func SaveCredentials(creds *Credentials, configDir string) error {
	if err := config.EnsureConfigDir(configDir); err != nil {
		return err
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	key := deriveKey()
	encrypted, err := encrypt(data, key)
	if err != nil {
		return err
	}

	credsPath := config.GetCredentialsPath(configDir)
	return fileutil.WriteFileAtomic(credsPath, encrypted, 0600)
}

// DeleteCredentials removes stored credentials.
// Returns nil if the file was removed or did not exist (idempotent).
func DeleteCredentials(configDir string) error {
	credsPath := config.GetCredentialsPath(configDir)
	if err := os.Remove(credsPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

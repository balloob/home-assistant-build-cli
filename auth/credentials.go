package auth

import (
	"encoding/json"
	"os"
	"time"

	"github.com/home-assistant/hab/config"
)

// Credentials stores authentication information
type Credentials struct {
	URL          string  `json:"url"`
	ClientID     string  `json:"client_id,omitempty"`
	AccessToken  string  `json:"access_token,omitempty"`
	RefreshToken string  `json:"refresh_token,omitempty"`
	TokenExpiry  float64 `json:"token_expiry,omitempty"`
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
	// Refresh if within 5 minutes of expiry
	return float64(time.Now().Unix()) >= (c.TokenExpiry - 300)
}

// LoadCredentials loads credentials from storage
func LoadCredentials(configDir string) (*Credentials, error) {
	// First check environment variables
	envURL := os.Getenv("HAB_URL")
	envToken := os.Getenv("HAB_TOKEN")

	if envURL != "" && envToken != "" {
		return &Credentials{
			URL:         envURL,
			AccessToken: envToken,
		}, nil
	}

	// Check for refresh token in environment
	envRefresh := os.Getenv("HAB_REFRESH_TOKEN")
	if envURL != "" && envRefresh != "" {
		return &Credentials{
			URL:          envURL,
			RefreshToken: envRefresh,
		}, nil
	}

	// Load from encrypted file
	credsPath := config.GetCredentialsPath(configDir)

	data, err := os.ReadFile(credsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Decrypt
	key := deriveKey()
	decrypted, err := decrypt(data, key)
	if err != nil {
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(decrypted, &creds); err != nil {
		return nil, err
	}

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
	if err := os.WriteFile(credsPath, encrypted, 0600); err != nil {
		return err
	}

	return nil
}

// DeleteCredentials removes stored credentials
func DeleteCredentials(configDir string) bool {
	credsPath := config.GetCredentialsPath(configDir)
	if err := os.Remove(credsPath); err != nil {
		return false
	}
	return true
}

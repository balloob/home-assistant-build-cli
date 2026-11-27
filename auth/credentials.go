package auth

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
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

// credentialsFile is the on-disk format. The URL is not a secret and stays
// readable so a login that can no longer be decrypted can still be identified.
type credentialsFile struct {
	URL  string `json:"url"`
	Data []byte `json:"data"`
}

var displayURLPattern = regexp.MustCompile(`^https?://[A-Za-z0-9._\-:\[\]]+(/[A-Za-z0-9._~%/\-]*)?$`)

// displayURL reduces a URL to what may be stored and printed in the clear:
// no userinfo, query or fragment, and nothing a shell would interpret, because
// the unencrypted part of the file is not authenticated and ends up in a
// suggested command. It returns "" when the URL does not qualify.
func displayURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	u.User, u.RawQuery, u.Fragment = nil, "", ""
	if s := u.String(); displayURLPattern.MatchString(s) {
		return s
	}
	return ""
}

// UnreadableCredentialsError is returned when the saved login exists but
// cannot be decrypted, e.g. because the machine identity changed.
type UnreadableCredentialsError struct {
	URL string // empty when the file predates the readable URL
	Err error
}

func (e *UnreadableCredentialsError) Error() string {
	target := ""
	if e.URL != "" {
		target = " for " + e.URL
	}
	return fmt.Sprintf("The saved login%s can't be read on this machine anymore. Log in again: %s", target, e.LoginCommand())
}

func (e *UnreadableCredentialsError) Unwrap() error { return e.Err }

// LoginCommand returns the command that replaces the unreadable login.
func (e *UnreadableCredentialsError) LoginCommand() string {
	if e.URL == "" {
		return "hab auth login"
	}
	return "hab auth login --url " + e.URL
}

// LoadCredentials loads credentials from storage
func LoadCredentials(configDir string) (*Credentials, error) {
	// Check environment variables
	envURL := os.Getenv("HAB_URL")
	envToken := os.Getenv("HAB_TOKEN")

	if envURL != "" && envToken != "" {
		return &Credentials{
			URL:         envURL,
			AccessToken: envToken,
			Source:      "env_token",
		}, nil
	}

	// Check for refresh token in environment
	envRefresh := os.Getenv("HAB_REFRESH_TOKEN")
	if envURL != "" && envRefresh != "" {
		return &Credentials{
			URL:          envURL,
			RefreshToken: envRefresh,
			Source:       "env_refresh_token",
		}, nil
	}

	// Load from encrypted file
	credsPath := config.GetCredentialsPath(configDir)

	data, err := os.ReadFile(credsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Fall through to supervisor check
		} else {
			return nil, err
		}
	} else {
		// Files written before the URL was stored in the clear are bare ciphertext.
		var file credentialsFile
		ciphertext := data
		if json.Unmarshal(data, &file) == nil && len(file.Data) > 0 {
			ciphertext = file.Data
		}

		decrypted, err := decrypt(ciphertext, deriveKey())
		migrate := len(file.Data) == 0
		if err != nil {
			decrypted, err = decrypt(ciphertext, legacyKey())
			migrate = true
		}
		if err != nil {
			return nil, &UnreadableCredentialsError{URL: displayURL(file.URL), Err: err}
		}

		var creds Credentials
		if err := json.Unmarshal(decrypted, &creds); err != nil {
			return nil, err
		}

		if migrate {
			// Best effort: the old file still works until the MAC changes.
			_ = SaveCredentials(&creds, configDir)
		}

		creds.Source = "credentials_file"
		return &creds, nil
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

	file, err := json.Marshal(credentialsFile{URL: displayURL(creds.URL), Data: encrypted})
	if err != nil {
		return err
	}

	credsPath := config.GetCredentialsPath(configDir)
	return fileutil.WriteFileAtomic(credsPath, file, 0600)
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

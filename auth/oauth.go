package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
)

const (
	// AuthorizePath is the OAuth authorization endpoint
	AuthorizePath = "/auth/authorize"
	// TokenPath is the OAuth token endpoint
	TokenPath = "/auth/token"
)

const (
	// tokenRequestTimeout is the HTTP timeout for OAuth token requests.
	tokenRequestTimeout = 30 * time.Second
)

// tokenHTTPClient is the HTTP client used for OAuth token requests.
// Using a dedicated client (rather than http.DefaultClient) ensures a
// sensible timeout and avoids leaking connections on slow networks.
var tokenHTTPClient = &http.Client{Timeout: tokenRequestTimeout}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// postTokenRequest posts form data to the token endpoint and decodes the response.
func postTokenRequest(tokenURL string, data url.Values) (*TokenResponse, error) {
	resp, err := tokenHTTPClient.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	return &tokenResp, nil
}

// computeExpiry returns the token expiry timestamp from the given ExpiresIn value.
// Falls back to defaultTokenExpiry when the server omits the field.
func computeExpiry(expiresIn int) float64 {
	if expiresIn == 0 {
		expiresIn = defaultTokenExpiry
	}
	return float64(time.Now().Unix() + int64(expiresIn))
}

// generateState generates a random state string for CSRF protection
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ExchangeCodeForTokens exchanges an authorization code for tokens
func ExchangeCodeForTokens(haURL, code, redirectURI string) (*Credentials, error) {
	tokenURL := strings.TrimRight(haURL, "/") + TokenPath
	clientID := GetClientID(redirectURI)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("redirect_uri", redirectURI)

	tokenResp, err := postTokenRequest(tokenURL, data)
	if err != nil {
		return nil, err
	}

	return &Credentials{
		URL:          haURL,
		ClientID:     clientID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenExpiry:  computeExpiry(tokenResp.ExpiresIn),
	}, nil
}

// RefreshAccessToken refreshes an expired access token
func RefreshAccessToken(creds *Credentials) (*Credentials, error) {
	tokenURL := strings.TrimRight(creds.URL, "/") + TokenPath

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", creds.RefreshToken)
	data.Set("client_id", creds.ClientID)

	tokenResp, err := postTokenRequest(tokenURL, data)
	if err != nil {
		return nil, err
	}

	// Use existing refresh token if new one not provided
	refreshToken := tokenResp.RefreshToken
	if refreshToken == "" {
		refreshToken = creds.RefreshToken
	}

	return &Credentials{
		URL:          creds.URL,
		ClientID:     creds.ClientID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: refreshToken,
		TokenExpiry:  computeExpiry(tokenResp.ExpiresIn),
	}, nil
}

// RunOAuthFlow runs the full OAuth flow
func RunOAuthFlow(haURL string) (*Credentials, error) {
	// Generate state for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Start callback server
	server := NewOAuthCallbackServer()
	redirectURI, err := server.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer server.Stop()

	// Build authorization URL
	authURL := BuildAuthorizeURL(strings.TrimRight(haURL, "/"), redirectURI, state)

	fmt.Println("\nOpening browser for authentication...")
	fmt.Printf("If browser doesn't open, visit: %s\n\n", authURL)

	// Open browser
	if err := browser.OpenURL(authURL); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}

	// Wait for callback
	result, err := server.WaitForCallback(5 * time.Minute)
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("OAuth error: %s", result.Error)
	}

	if result.State != state {
		return nil, fmt.Errorf("state mismatch - possible CSRF attack")
	}

	if result.Code == "" {
		return nil, fmt.Errorf("no authorization code received")
	}

	// Exchange code for tokens
	return ExchangeCodeForTokens(haURL, result.Code, redirectURI)
}

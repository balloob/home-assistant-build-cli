package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// CallbackResult contains the OAuth callback parameters
type CallbackResult struct {
	Code  string
	State string
	Error string
}

// OAuthCallbackServer handles OAuth callbacks
type OAuthCallbackServer struct {
	IP       string
	Port     int
	server   *http.Server
	result   *CallbackResult
	resultCh chan struct{}
	mu       sync.Mutex
}

// NewOAuthCallbackServer creates a new callback server
func NewOAuthCallbackServer() *OAuthCallbackServer {
	return &OAuthCallbackServer{
		resultCh: make(chan struct{}),
	}
}

// Start starts the callback server and returns the redirect URI
func (s *OAuthCallbackServer) Start() (string, error) {
	// Get local IP
	s.IP = getLocalIP()

	// Find available port
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:0", s.IP))
	if err != nil {
		return "", fmt.Errorf("failed to listen: %w", err)
	}
	s.Port = listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)

	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		s.server.Serve(listener)
	}()

	return fmt.Sprintf("http://%s:%d/callback", s.IP, s.Port), nil
}

func (s *OAuthCallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	s.mu.Lock()
	s.result = &CallbackResult{
		Code:  query.Get("code"),
		State: query.Get("state"),
		Error: query.Get("error"),
	}
	s.mu.Unlock()

	// Send response
	w.Header().Set("Content-Type", "text/html")
	if s.result.Error != "" {
		fmt.Fprintf(w, `
			<html>
			<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
				<h1>Authentication Failed</h1>
				<p>Error: %s</p>
				<p>You can close this window.</p>
			</body>
			</html>
		`, s.result.Error)
	} else {
		fmt.Fprint(w, `
			<html>
			<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
				<h1>Authentication Successful!</h1>
				<p>You can close this window and return to the terminal.</p>
			</body>
			</html>
		`)
	}

	// Signal that we have a result
	close(s.resultCh)
}

// WaitForCallback waits for the OAuth callback
func (s *OAuthCallbackServer) WaitForCallback(timeout time.Duration) (*CallbackResult, error) {
	select {
	case <-s.resultCh:
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("OAuth callback not received within timeout")
	}
}

// Stop stops the callback server
func (s *OAuthCallbackServer) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// getLocalIP returns the local network IP address
func getLocalIP() string {
	// Create a UDP connection to determine the local IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// BuildAuthorizeURL constructs the OAuth authorization URL
func BuildAuthorizeURL(haURL, redirectURI, state string) string {
	// Client ID must be the origin of the redirect URI
	clientID := GetClientID(redirectURI)

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)

	return fmt.Sprintf("%s/auth/authorize?%s", haURL, params.Encode())
}

// GetClientID extracts the origin (scheme://host:port) from a redirect URI to use as client ID
func GetClientID(redirectURI string) string {
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
}

package client

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// tlsConfig returns a *tls.Config that skips certificate verification
// when verifySSL is false, or nil (use Go defaults) when true.
func tlsConfig(verifySSL bool) *tls.Config {
	if verifySSL {
		return nil
	}
	return &tls.Config{InsecureSkipVerify: true}
}

// BuildWebSocketURL converts an HTTP URL to WebSocket URL
func BuildWebSocketURL(baseURL string) (string, error) {
	base := strings.TrimRight(baseURL, "/")

	// Convert scheme
	base = strings.Replace(base, "https://", "wss://", 1)
	base = strings.Replace(base, "http://", "ws://", 1)

	// Ensure scheme
	if !strings.Contains(base, "://") {
		base = "ws://" + base
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Supervisor proxy exposes WebSocket at /core/websocket instead of /api/websocket
	if parsed.Hostname() == "supervisor" {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/websocket"
	} else {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/api/websocket"
	}

	log.WithField("websocket_url", parsed.String()).Debug("Built WebSocket URL")

	return parsed.String(), nil
}


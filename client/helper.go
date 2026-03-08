package client

import (
	"fmt"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// BuildURL constructs an API URL from base URL and endpoint
func BuildURL(baseURL, endpoint string) (string, error) {
	base := strings.TrimRight(baseURL, "/")

	// Ensure scheme
	if !strings.Contains(base, "://") {
		base = "http://" + base
	}

	// Build full URL
	fullURL := fmt.Sprintf("%s/api/%s", base, strings.TrimLeft(endpoint, "/"))

	parsed, err := url.Parse(fullURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	log.WithFields(log.Fields{
		"base":     baseURL,
		"endpoint": endpoint,
		"result":   parsed.String(),
	}).Debug("Built URL")

	return parsed.String(), nil
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


package client

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
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

	fullURL := fmt.Sprintf("%s/api/websocket", base)

	parsed, err := url.Parse(fullURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	log.WithField("websocket_url", parsed.String()).Debug("Built WebSocket URL")

	return parsed.String(), nil
}

// PrintError prints an error message to stderr
func PrintError(err error) {
	PrintErrorString(err.Error())
}

// PrintErrorString prints an error message to stderr
func PrintErrorString(msg string) {
	if term.IsTerminal(int(os.Stderr.Fd())) {
		fmt.Fprintf(os.Stderr, "\033[1;31mError:\033[0m %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}
}

// PrintWarning prints a warning message to stderr
func PrintWarning(msg string) {
	if term.IsTerminal(int(os.Stderr.Fd())) {
		fmt.Fprintf(os.Stderr, "\033[1;33mWarning:\033[0m %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", msg)
	}
}

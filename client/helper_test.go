package client

import "testing"

func TestBuildWebSocketURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "standard HTTP",
			baseURL: "http://localhost:8123",
			want:    "ws://localhost:8123/api/websocket",
		},
		{
			name:    "standard HTTPS",
			baseURL: "https://my-ha.example.com",
			want:    "wss://my-ha.example.com/api/websocket",
		},
		{
			name:    "trailing slash",
			baseURL: "http://localhost:8123/",
			want:    "ws://localhost:8123/api/websocket",
		},
		{
			name:    "supervisor proxy",
			baseURL: "http://supervisor/core",
			want:    "ws://supervisor/core/websocket",
		},
		{
			name:    "supervisor proxy trailing slash",
			baseURL: "http://supervisor/core/",
			want:    "ws://supervisor/core/websocket",
		},
		{
			name:    "no scheme",
			baseURL: "localhost:8123",
			want:    "ws://localhost:8123/api/websocket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildWebSocketURL(tt.baseURL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("BuildWebSocketURL(%q) = %q, want %q", tt.baseURL, got, tt.want)
			}
		})
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
		want     string
	}{
		{
			name:     "standard",
			baseURL:  "http://localhost:8123",
			endpoint: "config",
			want:     "http://localhost:8123/api/config",
		},
		{
			name:     "supervisor proxy",
			baseURL:  "http://supervisor/core",
			endpoint: "config",
			want:     "http://supervisor/core/api/config",
		},
		{
			name:     "trailing slash on base",
			baseURL:  "http://localhost:8123/",
			endpoint: "states",
			want:     "http://localhost:8123/api/states",
		},
		{
			name:     "leading slash on endpoint",
			baseURL:  "http://localhost:8123",
			endpoint: "/config",
			want:     "http://localhost:8123/api/config",
		},
		{
			name:     "no scheme adds http",
			baseURL:  "192.168.1.100:8123",
			endpoint: "config",
			want:     "http://192.168.1.100:8123/api/config",
		},
		{
			name:     "https scheme",
			baseURL:  "https://my-ha.duckdns.org",
			endpoint: "services",
			want:     "https://my-ha.duckdns.org/api/services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildURL(tt.baseURL, tt.endpoint)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("BuildURL(%q, %q) = %q, want %q", tt.baseURL, tt.endpoint, got, tt.want)
			}
		})
	}
}

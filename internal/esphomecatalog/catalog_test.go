package esphomecatalog

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizePackageImportURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "github blob URL",
			input: "https://github.com/athom-tech/esp32-configs/blob/main/athom-smart-plug.yaml",
			want:  "github://athom-tech/esp32-configs/athom-smart-plug.yaml@main",
		},
		{
			name:  "raw github URL",
			input: "https://raw.githubusercontent.com/athom-tech/esp32-configs/main/athom-smart-plug.yaml",
			want:  "github://athom-tech/esp32-configs/athom-smart-plug.yaml@main",
		},
		{
			name:  "github package URL stays unchanged",
			input: "github://athom-tech/esp32-configs/athom-smart-plug.yaml@main",
			want:  "github://athom-tech/esp32-configs/athom-smart-plug.yaml@main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePackageImportURL(tt.input)
			if got != tt.want {
				t.Fatalf("normalizePackageImportURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeLinkedYAMLURL(t *testing.T) {
	got := normalizeLinkedYAMLURL("http://catalog.test", "https://github.com/esphome/esphome-devices/blob/main/src/docs/devices/dev-a/dev-a.yaml")
	want := "http://catalog.test/esphome/esphome-devices/main/src/docs/devices/dev-a/dev-a.yaml"
	if got != want {
		t.Fatalf("normalizeLinkedYAMLURL() = %q, want %q", got, want)
	}
}

func TestGetDeviceWithLinkedYAML(t *testing.T) {
	server := mockCatalogServer(t)
	defer server.Close()

	client := NewClient("main")
	client.apiRoot = server.URL
	client.rawRoot = server.URL

	device, err := client.GetDevice(context.Background(), "dev-a")
	if err != nil {
		t.Fatalf("GetDevice: %v", err)
	}
	if device.Title != "Device A" {
		t.Fatalf("Title = %q, want Device A", device.Title)
	}
	if device.ContentSource != "yaml_file" {
		t.Fatalf("ContentSource = %q, want yaml_file", device.ContentSource)
	}
	if device.YAMLContent == "" {
		t.Fatal("expected YAMLContent")
	}
}

func TestGetDeviceWithYAMLBlockFallback(t *testing.T) {
	server := mockCatalogServer(t)
	defer server.Close()

	client := NewClient("main")
	client.apiRoot = server.URL
	client.rawRoot = server.URL

	device, err := client.GetDevice(context.Background(), "dev-b")
	if err != nil {
		t.Fatalf("GetDevice: %v", err)
	}
	if device.ContentSource != "yaml_block" {
		t.Fatalf("ContentSource = %q, want yaml_block", device.ContentSource)
	}
	if device.YAMLContent == "" {
		t.Fatal("expected YAMLContent")
	}
}

func TestSearchFilters(t *testing.T) {
	server := mockCatalogServer(t)
	defer server.Close()

	client := NewClient("main")
	client.apiRoot = server.URL
	client.rawRoot = server.URL

	results, err := client.Search(context.Background(), SearchOptions{
		Query:      "dev",
		Board:      "esp32",
		DeviceType: "misc",
		Difficulty: 2,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Slug != "dev-a" {
		t.Fatalf("slug = %q, want dev-a", results[0].Slug)
	}
}

func TestDiscoverGitHubTokenFromEnvironment(t *testing.T) {
	t.Setenv("GH_TOKEN", "env-token")
	if got := discoverGitHubToken(); got != "env-token" {
		t.Fatalf("discoverGitHubToken() = %q, want env-token", got)
	}
}

func TestDiscoverGitHubTokenFromGHCLI(t *testing.T) {
	origLookPath := lookPath
	origCmd := ghTokenCommand
	defer func() {
		lookPath = origLookPath
		ghTokenCommand = origCmd
	}()
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	lookPath = func(file string) (string, error) { return "/usr/bin/gh", nil }
	ghTokenCommand = func(ctx context.Context, path string) ([]byte, error) { return []byte("gh-cli-token\n"), nil }

	if got := discoverGitHubToken(); got != "gh-cli-token" {
		t.Fatalf("discoverGitHubToken() = %q, want gh-cli-token", got)
	}
}

func TestFormatGitHubHTTPErrorRateLimit(t *testing.T) {
	err := formatGitHubHTTPError("catalog API", http.StatusForbidden, []byte(`{"message":"API rate limit exceeded"}`), false)
	if err == nil || !contains(err.Error(), "GH_TOKEN") {
		t.Fatalf("expected rate limit guidance error, got %v", err)
	}
}

func contains(value, sub string) bool {
	return strings.Contains(value, sub)
}

func mockCatalogServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/esphome/esphome-devices/contents/src/docs/devices":
			_, _ = io.WriteString(w, `[
				{"name":"dev-a","path":"src/docs/devices/dev-a","type":"dir"},
				{"name":"dev-b","path":"src/docs/devices/dev-b","type":"dir"}
			]`)
			return
		case "/esphome/esphome-devices/main/src/docs/devices/dev-a/index.md":
			_, _ = io.WriteString(w, "---\ntitle: Device A\nboard: esp32\ntype: misc\ndifficulty: 2\nproject-url: https://example.com/a\n---\n\nSee [config](dev-a.yaml).\n")
			return
		case "/esphome/esphome-devices/main/src/docs/devices/dev-a/dev-a.yaml":
			_, _ = io.WriteString(w, "esphome:\n  name: dev_a\n")
			return
		case "/esphome/esphome-devices/main/src/docs/devices/dev-b/index.md":
			_, _ = io.WriteString(w, "---\ntitle: Device B\nboard: esp8266\ntype: switch\ndifficulty: 1\n---\n\n```yaml\nesphome:\n  name: dev_b\n```\n")
			return
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, "not found")
			return
		}
	}))
}

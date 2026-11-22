package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadFile(t *testing.T) {
	for _, mode := range []string{"success", "existing", "truncated", "unauthorized", "redirect"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/api/backup/download/id" || r.URL.Query().Get("agent_id") != "local" || r.Header.Get("Authorization") != "Bearer token" {
					t.Errorf("request %s", r.URL.Path)
				}
				switch mode {
				case "unauthorized":
					w.WriteHeader(401)
				case "redirect":
					w.Header().Set("Location", "/other")
					w.WriteHeader(302)
				case "truncated":
					w.Header().Set("Content-Length", "100")
				}
				if _, err := fmt.Fprint(w, "archive"); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			dir := t.TempDir()
			destination := filepath.Join(dir, "backup.tar")
			if mode == "existing" {
				if err := os.WriteFile(destination, []byte("original"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			c := NewRestClient(server.URL, "token")
			size, err := c.DownloadFile(context.Background(), "backup/download/id", url.Values{"agent_id": {"local"}}, destination)
			if mode == "success" {
				if err != nil || size != 7 {
					t.Fatalf("%d %v", size, err)
				}
				data, err := os.ReadFile(destination)
				if err != nil || string(data) != "archive" {
					t.Fatalf("%s %v", data, err)
				}
			} else {
				if err == nil {
					t.Fatal("expected failure")
				}
				if mode == "existing" {
					data, err := os.ReadFile(destination)
					if err != nil || string(data) != "original" || calls != 0 {
						t.Fatal("existing file changed or network used")
					}
				} else if _, err := os.Stat(destination); !os.IsNotExist(err) {
					t.Fatalf("partial file published: %v", err)
				}
			}
			files, err := filepath.Glob(filepath.Join(dir, ".hab-download-*"))
			if err != nil || len(files) != 0 {
				t.Fatalf("temporary files left: %v %v", files, err)
			}
			if mode == "redirect" && calls != 1 {
				t.Fatal("redirect followed")
			}
		})
	}
}

func TestUploadFile(t *testing.T) {
	for _, kind := range []string{"backup", "media"} {
		t.Run(kind, func(t *testing.T) {
			content := strings.Repeat("binary\x00", 10000)
			filename := "archive.tar"
			if kind == "media" {
				filename = "song.mp3"
			}
			file := filepath.Join(t.TempDir(), filename)
			if err := os.WriteFile(file, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer token" {
					t.Error("method or auth")
				}
				if kind == "backup" && len(r.URL.Query()["agent_id"]) != 2 {
					t.Error("missing agents")
				}
				reader, err := r.MultipartReader()
				if err != nil {
					t.Error(err)
					return
				}
				part, err := reader.NextPart()
				if err != nil {
					t.Error(err)
					return
				}
				if kind == "media" {
					data, err := io.ReadAll(part)
					if err != nil || part.FormName() != "media_content_id" || string(data) != "media-source://media_source/local" {
						t.Error("media folder missing")
					}
					part, err = reader.NextPart()
					if err != nil {
						t.Error(err)
						return
					}
				}
				if kind == "media" && !strings.HasPrefix(part.Header.Get("Content-Type"), "audio/") {
					t.Errorf("media content type: %s", part.Header.Get("Content-Type"))
				}
				data, err := io.ReadAll(part)
				if err != nil || part.FormName() != "file" || part.FileName() != filename || string(data) != content {
					t.Error("file body differs")
				}
				if _, err := reader.NextPart(); err != io.EOF {
					t.Errorf("trailing multipart %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(201)
				if _, err := fmt.Fprint(w, `{"backup_id":"backup"}`); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			fields := map[string]string{}
			if kind == "media" {
				fields["media_content_id"] = "media-source://media_source/local"
			}
			c := NewRestClient(server.URL, "token")
			result, err := c.UploadFile(context.Background(), kind+"/upload", url.Values{"agent_id": {"local", "cloud"}}, file, fields)
			if err != nil || result.(map[string]any)["backup_id"] != "backup" {
				t.Fatalf("%#v %v", result, err)
			}
		})
	}
}

func TestTransferCancellationAndEarlyRejection(t *testing.T) {
	file := filepath.Join(t.TempDir(), "large.tar")
	if err := os.WriteFile(file, make([]byte, 4<<20), 0600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(403) }))
	defer server.Close()
	c := NewRestClient(server.URL, "token")
	c.Timeout = time.Second
	if _, err := c.UploadFile(context.Background(), "backup/upload", nil, file, nil); err == nil {
		t.Fatal("rejection accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.DownloadFile(ctx, "backup/download/id", url.Values{"password": {"private-marker"}}, filepath.Join(t.TempDir(), "file"))
	if err == nil || strings.Contains(err.Error(), "private-marker") {
		t.Fatalf("error exposes query: %v", err)
	}
}

func TestManagementRESTForms(t *testing.T) {
	methods := []string{"GET", "POST", "DELETE"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method || r.URL.Path != "/api/config/config_entries/flow/flow" {
					t.Error("unexpected route")
				}
				if method == "POST" {
					var data map[string]any
					if err := json.NewDecoder(r.Body).Decode(&data); err != nil || data == nil || len(data) != 0 {
						t.Errorf("empty form lost: %v %v", data, err)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if _, err := fmt.Fprint(w, `{"type":"create_entry","result":{"entry_id":"entry"}}`); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			c := NewRestClient(server.URL, "token")
			var body any
			if method == "POST" {
				body = map[string]any{}
			}
			result, err := c.RequestContext(context.Background(), method, "config/config_entries/flow/flow", body)
			if err != nil || result.(map[string]any)["type"] != "create_entry" {
				t.Fatalf("%v %v", result, err)
			}
		})
	}
}

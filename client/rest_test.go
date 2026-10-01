package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsJSONContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"exact match", "application/json", true},
		{"with charset", "application/json; charset=utf-8", true},
		{"text html", "text/html", false},
		{"text plain", "text/plain", false},
		{"empty string", "", false},
		{"partial match prefix", "application/javascript", false},
		{"case as-is", "application/json;charset=UTF-8", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isJSONContentType(tt.contentType)
			if got != tt.want {
				t.Errorf("isJSONContentType(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{Code: "NOT_FOUND", Message: "Resource not found"}

	if err.Error() != "Resource not found" {
		t.Errorf("APIError.Error() = %q, want %q", err.Error(), "Resource not found")
	}

	// Verify it implements the error interface
	var _ error = err
}

func TestNewRestClient(t *testing.T) {
	rc := NewRestClient("http://localhost:8123", "test-token")

	if rc.BaseURL != "http://localhost:8123" {
		t.Errorf("BaseURL = %q, want %q", rc.BaseURL, "http://localhost:8123")
	}
	if rc.Token != "test-token" {
		t.Errorf("Token = %q, want %q", rc.Token, "test-token")
	}
	if rc.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", rc.Timeout, DefaultTimeout)
	}
	if !rc.VerifySSL {
		t.Error("VerifySSL should default to true")
	}
}

func TestNewRestClientWithOptions(t *testing.T) {
	rc := NewRestClientWithOptions("http://localhost:8123", "tok", 60_000_000_000, false)

	if rc.BaseURL != "http://localhost:8123" {
		t.Errorf("BaseURL = %q, want %q", rc.BaseURL, "http://localhost:8123")
	}
	if rc.Token != "tok" {
		t.Errorf("Token = %q, want %q", rc.Token, "tok")
	}
	if rc.Timeout != 60_000_000_000 {
		t.Errorf("Timeout = %v, want 60s", rc.Timeout)
	}
	if rc.VerifySSL {
		t.Error("VerifySSL should be false")
	}
}

func TestCallServiceWithResponseUsesQueryParam(t *testing.T) {
	var gotPath, gotQuery string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"changed_states":[],"service_response":{}}`))
	}))
	defer srv.Close()

	rc := NewRestClient(srv.URL, "tok")
	if _, err := rc.CallServiceWithResponse("weather", "get_forecasts", map[string]interface{}{"type": "daily"}); err != nil {
		t.Fatalf("CallServiceWithResponse: %v", err)
	}
	if gotPath != "/api/services/weather/get_forecasts" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "return_response" {
		t.Errorf("query = %q, want return_response", gotQuery)
	}
	if _, ok := gotBody["return_response"]; ok {
		t.Error("body must not contain return_response")
	}
	if gotBody["type"] != "daily" {
		t.Errorf("body type = %v, want daily", gotBody["type"])
	}
}

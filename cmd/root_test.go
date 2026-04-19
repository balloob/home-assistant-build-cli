package cmd

import (
	"context"
	"testing"

	"github.com/home-assistant/hab/client"
)

func TestExecutableName(t *testing.T) {
	tests := []struct {
		name string
		arg0 string
		want string
	}{
		{name: "empty arg", arg0: "", want: "hab"},
		{name: "unix path", arg0: "/usr/local/bin/hab", want: "hab"},
		{name: "windows path", arg0: `C:\Program Files\hab\hab.exe`, want: "hab.exe"},
		{name: "bare name", arg0: "hab", want: "hab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := executableName(tt.arg0); got != tt.want {
				t.Errorf("executableName(%q) = %q, want %q", tt.arg0, got, tt.want)
			}
		})
	}
}

func TestClassifyErrorIncludesDetails(t *testing.T) {
	err := &client.APIError{
		Code:              client.ErrCodeValidationError,
		Message:           "validation failed",
		Details:           map[string]any{"line": 12},
		Category:          "validation",
		SuggestedFix:      "Fix payload",
		SuggestedCommands: []string{"hab action docs light.turn_on --json"},
	}

	code, msg, details := classifyError(err)
	if code != client.ErrCodeValidationError {
		t.Fatalf("code = %q, want %q", code, client.ErrCodeValidationError)
	}
	if msg != "validation failed" {
		t.Fatalf("msg = %q, want validation failed", msg)
	}
	if details["line"] != 12 {
		t.Fatalf("details = %#v, want line 12", details)
	}
	if details["category"] != "validation" {
		t.Fatalf("details = %#v, want category validation", details)
	}
	if details["suggested_fix"] != "Fix payload" {
		t.Fatalf("details = %#v, want suggested_fix", details)
	}
}

func TestClassifyErrorTimeout(t *testing.T) {
	code, msg, details := classifyError(context.DeadlineExceeded)
	if code != client.ErrCodeTimeout {
		t.Fatalf("code = %q, want %q", code, client.ErrCodeTimeout)
	}
	if msg == "" {
		t.Fatal("expected non-empty timeout message")
	}
	if details["retryable"] != true {
		t.Fatalf("details = %#v, want retryable true", details)
	}
}

func TestClassifyCancelledAPIError(t *testing.T) {
	err := client.NewCancelledError("delete area")
	code, msg, details := classifyError(err)
	if code != client.ErrCodeCancelled {
		t.Fatalf("code = %q, want %q", code, client.ErrCodeCancelled)
	}
	if msg == "" {
		t.Fatal("expected non-empty cancelled message")
	}
	if details["category"] != "cancellation" {
		t.Fatalf("details = %#v, want cancellation category", details)
	}
}

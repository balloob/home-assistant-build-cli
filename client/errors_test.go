package client

import "testing"

func TestAPIErrorDetailsMapIncludesStructuredFields(t *testing.T) {
	err := &APIError{
		Code:                ErrCodeValidationError,
		Message:             "validation failed",
		Details:             map[string]any{"field": "name"},
		Category:            "validation",
		Retryable:           false,
		LikelyCause:         "invalid payload",
		SuggestedFix:        "fix payload",
		SuggestedCommands:   []string{"hab action docs light.turn_on --json"},
		PrerequisiteMissing: "required_field",
		Transport:           "rest",
		StatusCode:          400,
	}

	details := err.DetailsMap()
	if details["field"] != "name" {
		t.Fatalf("details = %#v, want field=name", details)
	}
	if details["category"] != "validation" {
		t.Fatalf("details = %#v, want category=validation", details)
	}
	if details["suggested_fix"] != "fix payload" {
		t.Fatalf("details = %#v, want suggested_fix", details)
	}
	if details["status_code"] != 400 {
		t.Fatalf("details = %#v, want status_code=400", details)
	}
}

func TestNewCancelledError(t *testing.T) {
	err := NewCancelledError("delete area")
	if err.Code != ErrCodeCancelled {
		t.Fatalf("code = %q, want %q", err.Code, ErrCodeCancelled)
	}
	if !err.Retryable {
		t.Fatal("expected retryable cancelled error")
	}
	details := err.DetailsMap()
	if details["action"] != "delete area" {
		t.Fatalf("details = %#v, want action", details)
	}
}

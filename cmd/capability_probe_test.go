package cmd

import (
	"testing"

	"github.com/home-assistant/hab/client"
)

func TestStatusFromErrorRestricted(t *testing.T) {
	status := statusFromError(&client.APIError{Code: client.ErrCodePermissionDenied, Message: "denied"})
	if !status.Available {
		t.Fatalf("available = %v, want true", status.Available)
	}
	if !status.Restricted {
		t.Fatalf("restricted = %v, want true", status.Restricted)
	}
}

func TestStatusFromErrorUnsupported(t *testing.T) {
	status := statusFromError(&client.APIError{Code: client.ErrCodeNotFound, Message: "not found"})
	if status.Available {
		t.Fatalf("available = %v, want false", status.Available)
	}
	if status.Reason != "unsupported" {
		t.Fatalf("reason = %q, want unsupported", status.Reason)
	}
}

func TestPlanRequested(t *testing.T) {
	if planRequested(false, false) {
		t.Fatal("expected false when both flags are unset")
	}
	if !planRequested(true, false) {
		t.Fatal("expected true when --plan is set")
	}
	if !planRequested(false, true) {
		t.Fatal("expected true when --dry-run is set")
	}
}

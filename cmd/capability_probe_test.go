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

type fakeCommander struct {
	cmdType string
	params  map[string]interface{}
	err     error
}

func (f *fakeCommander) SendCommand(cmdType string, params map[string]interface{}) (interface{}, error) {
	f.cmdType = cmdType
	f.params = params
	if f.err != nil {
		return nil, f.err
	}
	return map[string]interface{}{}, nil
}

func TestProbeSupervisorAvailable(t *testing.T) {
	ws := &fakeCommander{}
	status := probeSupervisor(ws)
	if !status.Available {
		t.Fatalf("available = %v, want true", status.Available)
	}
	if ws.cmdType != "supervisor/api" || ws.params["endpoint"] != "/info" || ws.params["method"] != "get" {
		t.Fatalf("sent %q %v, want supervisor/api with endpoint /info and method get", ws.cmdType, ws.params)
	}
}

func TestProbeSupervisorUnknownCommand(t *testing.T) {
	ws := &fakeCommander{err: &client.APIError{Code: "unknown_command", Message: "Unknown command."}}
	status := probeSupervisor(ws)
	if status.Available {
		t.Fatalf("available = %v, want false", status.Available)
	}
	if status.Reason != "supervisor endpoint unavailable" {
		t.Fatalf("reason = %q, want supervisor endpoint unavailable", status.Reason)
	}
}

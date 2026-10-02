package client

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/home-assistant/hab/internal/dashboardpatch"
)

type dashboardCommander struct {
	command string
	params  map[string]any
	err     error
}

func (c *dashboardCommander) SendCommandContext(ctx context.Context, command string, params map[string]any) (any, error) {
	c.command = command
	c.params = params
	return map[string]any{"views": []any{}}, c.err
}

func TestDashboardStoreAPIContract(t *testing.T) {
	for _, path := range []string{"lovelace", "my-dashboard"} {
		t.Run(path, func(t *testing.T) {
			ws := &dashboardCommander{}
			store := DashboardStore{WS: ws, URLPath: path}
			config, err := store.Load(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if ws.command != "lovelace/config" {
				t.Fatal(ws.command)
			}
			if err := store.Save(context.Background(), config); err != nil {
				t.Fatal(err)
			}
			if ws.command != "lovelace/config/save" || !reflect.DeepEqual(ws.params["config"], config) {
				t.Fatalf("incorrect save payload: %+v", ws)
			}
			_, hasURL := ws.params["url_path"]
			if hasURL != (path != "lovelace") {
				t.Fatal("incorrect default-dashboard mapping")
			}
			if _, ok := ws.params["if_match"]; ok {
				t.Fatal("invented conditional save API")
			}
		})
	}
}

func TestDashboardStoreDistinguishesRejectionFromUncertainty(t *testing.T) {
	for _, tc := range []struct {
		err      error
		rejected bool
	}{
		{&APIError{Code: "unauthorized", Category: "api"}, true},
		{&APIError{Code: "unknown_command", Category: "api"}, true},
		{&APIError{Code: "error", Category: "api"}, false},
		{&APIError{Code: ErrCodeTimeout, Category: "timeout"}, false},
		{context.Canceled, false},
	} {
		t.Run(tc.err.Error(), func(t *testing.T) {
			store := DashboardStore{WS: &dashboardCommander{err: tc.err}, URLPath: "test-dashboard"}
			err := store.Save(context.Background(), map[string]any{})
			var rejection *dashboardpatch.Rejected
			if errors.As(err, &rejection) != tc.rejected {
				t.Fatalf("wrong classification: %v", err)
			}
		})
	}
}

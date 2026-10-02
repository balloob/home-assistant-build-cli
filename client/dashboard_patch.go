package client

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/home-assistant/hab/internal/dashboardpatch"
)

// DashboardStore reuses one authenticated connection for the entire workflow.
type DashboardStore struct {
	WS interface {
		SendCommandContext(context.Context, string, map[string]any) (any, error)
	}
	URLPath string
}

func (s DashboardStore) params() map[string]any {
	params := map[string]any{}
	if s.URLPath != "lovelace" {
		params["url_path"] = s.URLPath
	}
	return params
}

func (s DashboardStore) Load(ctx context.Context) (map[string]any, error) {
	result, err := s.WS.SendCommandContext(ctx, "lovelace/config", s.params())
	if err != nil {
		return nil, err
	}
	config, ok := result.(map[string]any)
	if !ok || config == nil {
		return nil, fmt.Errorf("unexpected dashboard config response")
	}
	return config, nil
}

func (s DashboardStore) Save(ctx context.Context, config map[string]any) error {
	params := s.params()
	params["config"] = config
	_, err := s.WS.SendCommandContext(ctx, "lovelace/config/save", params)
	var apiErr *APIError
	// Only dispatch/permission/validation failures prove that save did not run.
	// An arbitrary server error may occur after persistence; reconcile it by read.
	if errors.As(err, &apiErr) && apiErr.Category == "api" && slices.Contains([]string{"unauthorized", "permission_denied", "unknown_command", "invalid_format", "config_not_found"}, strings.ToLower(apiErr.Code)) {
		return &dashboardpatch.Rejected{Err: err}
	}
	return err
}

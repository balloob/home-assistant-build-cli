package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/home-assistant/hab/internal/dashboardpatch"
)

func TestDashboardPatchInputAndRedaction(t *testing.T) {
	for _, tc := range []struct {
		name  string
		flags dashboardPatchFlags
	}{
		{"missing_revision", dashboardPatchFlags{set: []string{`/name="test"`}}},
		{"ambiguous_input", dashboardPatchFlags{plan: true, input: InputFlags{Data: `{}`, File: "secret.yaml"}}},
		{"invalid_json", dashboardPatchFlags{plan: true, set: []string{`/password=secret-invalid-json`}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.flags.timeout = time.Second
			tc.flags.diffLimit = 50
			_, err := parseDashboardPatch(&tc.flags)
			if err == nil {
				t.Fatal("invalid input accepted")
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("input value leaked into error")
			}
		})
	}
	result := dashboardpatch.Result{Changes: []dashboardpatch.Change{
		{Path: "/views/0/cards/0/access_token", Before: "oldsecret", After: "newsecret"},
		{Path: "/views/0/cards/0", After: map[string]any{"password": "embeddedsecret", "url": "https://user:urlsecret@example.com/?token=querysecret"}},
	}}
	redactDashboardResult(&result)
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "secret") {
		t.Fatalf("secret leaked: %s", data)
	}
}

func TestGenericPlanRedactsOpaqueInputs(t *testing.T) {
	cmd := newDashboardPatchCommand()
	if err := cmd.ParseFlags([]string{"--data", `{"password":"opaque-secret"}`, "--set", `/api_key="set-secret"`}); err != nil {
		t.Fatal(err)
	}
	values := changedFlagValues(cmd)
	data, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "secret") {
		t.Fatalf("generic plan leaked opaque data: %s", data)
	}
}

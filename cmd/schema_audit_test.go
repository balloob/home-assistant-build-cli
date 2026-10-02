package cmd

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestEveryExecutableCommandHasReviewedContract(t *testing.T) {
	inventory := auditedDataShapes()
	seen := map[string]bool{}
	walkCommandTree(rootCmd, func(cmd *cobra.Command) {
		if !cmd.Runnable() {
			return
		}
		path := schemaPath(cmd)
		seen[path] = true
		t.Run(path, func(t *testing.T) {
			if _, ok := inventory[path]; !ok {
				t.Fatal("executable missing from reviewed contract inventory")
			}
			s := buildSchemaTree(cmd)
			if expected := reviewedEffect(t, path, cmd.Name()); s.SideEffect != expected {
				t.Fatalf("side_effect=%s want=%s", s.SideEffect, expected)
			}
			if len(s.Transports) == 0 {
				t.Fatal("missing transport contract")
			}
			if !slices.Contains([]string{"read", "write", "destructive"}, s.SideEffect) {
				t.Fatalf("invalid side effect: %s", s.SideEffect)
			}
			if len(s.OutputContract.Variants) == 0 {
				t.Fatal("missing output variants")
			}
			for _, v := range s.OutputContract.Variants {
				if v.Data == nil || v.Data.Type == "" {
					t.Fatalf("no data shape for %s", v.Name)
				}
				if (v.Name == "brief" || v.Name == "count") && cmd.Flags().Lookup(v.Name) == nil {
					t.Fatalf("advertised unsupported variant: %s", v.Name)
				}
			}
			if s.SideEffect != "read" && s.Preview == "none" {
				t.Fatal("mutation has no preview")
			}
			for _, source := range s.InputSources {
				if source == "data" && cmd.Flags().Lookup("data") == nil {
					t.Fatal("advertised nonexistent input source")
				}
			}
			if slices.Contains(s.InputSources, "file") && (s.PayloadSchema == nil || s.PayloadSchema.Type != "object") {
				t.Fatal("missing structured object input contract")
			}
		})
	})
	for path := range inventory {
		if path == "hab help" {
			continue
		} // Cobra adds help on first execution.
		if !seen[path] {
			t.Errorf("audit entry has no executable: %s", path)
		}
	}
}

func reviewedEffect(t *testing.T, path, name string) string {
	t.Helper()
	if slices.Contains([]string{"hab network configure", "hab esphome serial erase-flash", "hab esphome context clear"}, path) {
		return "destructive"
	}
	if path == "hab marketplace custom remove" {
		return "write"
	}
	groups := map[string]string{
		"read":        "agents analyze boards config-check config-read data detect discover docs entities get guide health help history info items list lists logbook logs ports related releases removed render schema search show solar-forecast status trace types updates url validate version",
		"write":       "accept-warning acknowledge activate add assign build call clear-new complete config-patch config-write configure create create-from-blueprint disable dismiss enable fire github-connect ignore import install login patch probe refresh reload rename run save-config set set-beta set-preferred set-version uncomplete unignore update upload use",
		"destructive": "delete erase-flash logout remove restart restore uninstall",
	}
	if path == "hab capability probe" {
		return "read"
	}
	if path == "hab overview" {
		return "read"
	}
	for effect, names := range groups {
		if slices.Contains(strings.Fields(names), name) {
			return effect
		}
	}
	t.Fatalf("operation needs a side-effect review: %s", path)
	return ""
}

func TestCriticalContracts(t *testing.T) {
	for _, tc := range []struct{ path, effect, transport, shape string }{
		{"dashboard save-config", "write", "ws", "null"},
		{"scene activate", "write", "rest", "null"},
		{"esphome upload", "write", "esphome", "stream"},
		{"esphome config-patch", "write", "esphome", "object"},
		{"repairs ignore", "write", "ws", "null"},
		{"thread set-preferred", "write", "ws", "null"},
		{"automation trace", "read", "ws", "any"},
		{"entity history", "read", "rest", "array"},
		{"system logs", "read", "rest", "string"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			command, err := resolveSchemaTarget(strings.Fields(tc.path))
			if err != nil {
				t.Fatal(err)
			}
			s := buildSchemaTree(command)
			if s.SideEffect != tc.effect || !slices.Contains(s.Transports, tc.transport) {
				t.Fatalf("bad classification: %+v", s)
			}
			for _, variant := range s.OutputContract.Variants {
				if variant.Name == "full" && tc.shape != "stream" && variant.Data.Type != tc.shape {
					t.Fatalf("full shape=%s want=%s", variant.Data.Type, tc.shape)
				}
			}
		})
	}
}

func TestContextAwareInvocationContracts(t *testing.T) {
	for _, path := range []string{"esphome build", "esphome upload", "esphome run", "esphome logs"} {
		t.Run(path, func(t *testing.T) {
			cmd, err := resolveSchemaTarget(strings.Fields(path))
			if err != nil {
				t.Fatal(err)
			}
			s := buildSchemaTree(cmd)
			if len(s.Args) != 1 || s.Args[0].Required || !slices.Contains(s.InputSources, "saved_context") {
				t.Fatal("saved context was incorrectly described as a required argument")
			}
		})
	}
	s := buildSchemaTree(schemaCmd)
	if len(s.Args) != 1 || !s.Args[0].Variadic {
		t.Fatal("multiword schema path is not described")
	}
}

func TestSchemaStrictPathsAndCompactDiscovery(t *testing.T) {
	for _, path := range []string{"dashboard nonexistent", "dashboard get typo", "automation create extra", "no-such-command"} {
		if _, err := resolveSchemaTarget(strings.Fields(path)); err == nil {
			t.Fatalf("accepted unknown path: %s", path)
		}
	}
	index, err := buildCommandIndex(rootCmd, "dashboard patch", 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if index.Total != 1 || len(index.Commands) != 1 || !index.Complete || index.Commands[0].Path != "hab dashboard patch" {
		t.Fatalf("bad search: %+v", index)
	}
	page, err := buildCommandIndex(rootCmd, "dashboard", 0, 2)
	if err != nil || page.Complete || page.NextOffset == nil || *page.NextOffset != 2 {
		t.Fatalf("bad page: %+v %v", page, err)
	}
	command, err := resolveSchemaTarget([]string{"dashboard", "patch"})
	if err != nil {
		t.Fatal(err)
	}
	schema := buildSchemaTree(command)
	compact, err := compactCommandSchema(schema)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(compact)
	if err != nil {
		t.Fatal(err)
	}
	full, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "success_envelope") || strings.Contains(string(data), "partial_envelope") {
		t.Fatal("compact schema repeats envelope definitions")
	}
	if compact["schema_id"] == "" || compact["envelope_ref"] != envelopeSchemaVersion || len(data) >= len(full) {
		t.Fatal("missing identity or compaction")
	}
	if schema.OutputContract.Variants[0].Envelope == nil {
		t.Fatal("compact schema modified shared contracts")
	}
	t.Logf("dashboard patch schema: full=%d compact=%d bytes", len(full), len(data))
}

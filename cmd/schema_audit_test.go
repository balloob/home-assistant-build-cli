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
	management := map[string]string{
		"hab app logs":                        "read",
		"hab app rebuild":                     "destructive",
		"hab app refresh":                     "write",
		"hab app repository add":              "write",
		"hab app repository list":             "read",
		"hab app repository remove":           "destructive",
		"hab app security":                    "destructive",
		"hab app stats":                       "read",
		"hab app validate":                    "read",
		"hab application-credential config":   "read",
		"hab application-credential create":   "write",
		"hab application-credential delete":   "destructive",
		"hab application-credential entry":    "read",
		"hab application-credential list":     "read",
		"hab assist devices":                  "read",
		"hab assist exposure defaults get":    "read",
		"hab assist exposure defaults set":    "write",
		"hab assist exposure list":            "read",
		"hab assist exposure set":             "write",
		"hab assist languages":                "read",
		"hab assist pipeline create":          "write",
		"hab assist pipeline delete":          "destructive",
		"hab assist pipeline get":             "read",
		"hab assist pipeline list":            "read",
		"hab assist pipeline set-preferred":   "write",
		"hab assist pipeline update":          "write",
		"hab assist runs get":                 "read",
		"hab assist runs list":                "read",
		"hab backup download":                 "read",
		"hab backup upload":                   "write",
		"hab diagnostics download":            "read",
		"hab host datadisk list":              "read",
		"hab host datadisk move":              "destructive",
		"hab host disks list":                 "read",
		"hab host disks usage":                "read",
		"hab host info":                       "read",
		"hab host options":                    "write",
		"hab host os info":                    "read",
		"hab host os sync":                    "destructive",
		"hab host os update":                  "destructive",
		"hab host reboot":                     "destructive",
		"hab host shutdown":                   "destructive",
		"hab integration delete":              "destructive",
		"hab integration flow abort":          "write",
		"hab integration flow get":            "read",
		"hab integration flow list":           "read",
		"hab integration flow start":          "write",
		"hab integration flow step":           "write",
		"hab integration options abort":       "write",
		"hab integration options get":         "read",
		"hab integration options start":       "write",
		"hab integration options step":        "write",
		"hab integration subentry delete":     "destructive",
		"hab integration subentry flow abort": "write",
		"hab integration subentry flow get":   "read",
		"hab integration subentry flow start": "write",
		"hab integration subentry flow step":  "write",
		"hab integration subentry list":       "read",
		"hab integration subentry update":     "write",
		"hab integration types":               "read",
		"hab integration update":              "write",
		"hab media browse":                    "read",
		"hab media delete":                    "destructive",
		"hab media resolve":                   "read",
		"hab media search":                    "read",
		"hab media upload":                    "write",
		"hab mount create":                    "write",
		"hab mount delete":                    "destructive",
		"hab mount list":                      "read",
		"hab mount options":                   "write",
		"hab mount reload":                    "write",
		"hab mount update":                    "write",
		"hab network interfaces list":         "read",
		"hab network interfaces scan":         "read",
		"hab network interfaces update":       "destructive",
		"hab repairs flow abort":              "write",
		"hab repairs flow get":                "read",
		"hab repairs flow start":              "write",
		"hab repairs flow step":               "write",
		"hab repairs get":                     "read",
		"hab statistics adjust-sum":           "destructive",
		"hab statistics clear":                "destructive",
		"hab statistics metadata list":        "read",
		"hab statistics metadata update":      "write",
		"hab statistics query":                "read",
		"hab statistics refresh-issues":       "write",
		"hab statistics summary":              "read",
		"hab statistics validate":             "read",
		"hab system config detect":            "read",
		"hab system config get":               "read",
		"hab system config update":            "write",
		"hab user create":                     "write",
		"hab user credentials create":         "write",
		"hab user delete":                     "destructive",
		"hab user list":                       "read",
		"hab user password":                   "destructive",
		"hab user update":                     "write",
		"hab user username":                   "write",
	}
	if effect, ok := management[path]; ok {
		return effect
	}

	if slices.Contains([]string{"hab app stop", "hab network configure", "hab esphome serial erase-flash", "hab esphome context clear"}, path) {
		return "destructive"
	}
	if path == "hab marketplace custom remove" {
		return "write"
	}
	groups := map[string]string{
		"read":        "agents analyze boards config-check config-read data detect discover docs entities get guide health help history info items list lists logbook logs ports related releases removed render schema search show solar-forecast status trace types updates url validate version",
		"write":       "accept-warning acknowledge activate add assign build call clear-new complete config-patch config-write configure create create-from-blueprint disable dismiss enable fire github-connect ignore import install login patch probe refresh reload rename run save-config set set-beta set-preferred set-version start uncomplete unignore update upload use",
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

package cmd

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/internal/shortid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestShortIDCommandCoverage(t *testing.T) {
	// Include read, mutation, destructive, filtered discovery, and scoped inputs.
	for _, tt := range []struct {
		path, kind, flag string
		arg              int
	}{
		{"device get", "device", "", 0}, {"device entities", "device", "device-id", 0}, {"device update", "device", "", 0}, {"device delete", "device", "", 0},
		{"device list", "device", "device-id", 0}, {"entity list", "device", "device", 0},
		{"integration get", "config_entry", "", 0}, {"integration reload", "config_entry", "", 0}, {"integration enable", "config_entry", "", 0}, {"integration disable", "config_entry", "", 0},
		{"category update", "category", "", 0}, {"category delete", "category", "", 0}, {"category assign", "category", "category", 0},
		{"person create", "user", "user-id", 0}, {"person update", "person", "", 0}, {"person delete", "person", "", 0},
		{"area update", "label", "labels", 0}, {"floor update", "floor", "", 0}, {"label assign", "label", "", 0},
		{"thread get", "thread", "dataset", 0}, {"thread delete", "thread", "", 0}, {"thread set-preferred", "thread", "", 0},
		{"backup get", "backup", "", 0}, {"backup delete", "backup", "", 0}, {"backup restore", "backup", "", 0}, {"backup create", "backup_agent", "agent", 0},
		{"dashboard delete", "dashboard", "", 0}, {"dashboard update", "dashboard", "", 0},
		{"todo complete", "todo_item", "", 1}, {"todo uncomplete", "todo_item", "", 1}, {"todo update", "todo_item", "", 1}, {"todo remove", "todo_item", "", 1},
		{"calendar delete", "calendar_event", "", 1}, {"notification dismiss", "notification", "", 0}, {"repairs ignore", "repair", "", 1}, {"repairs unignore", "repair", "", 1},
		{"automation get", "automation", "", 0}, {"automation trigger create", "automation", "", 0}, {"automation trace", "trace", "run-id", 0},
		{"scene activate", "scene", "", 0}, {"script action update", "script", "", 0},
		{"helper template delete", "config_entry", "", 0}, {"helper input-boolean delete", "helper:input_boolean", "", 0},
	} {
		t.Run(tt.path+"/"+tt.flag, func(t *testing.T) {
			cmd, err := resolveSchemaTarget(strings.Fields(tt.path))
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, s := range shortIDSelectors(cmd, nil) {
				if s.Namespace.Kind == tt.kind && s.Flag == tt.flag && (s.Flag != "" || s.Argument == tt.arg) {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing selector %s: %+v", tt.kind, shortIDSelectors(cmd, nil))
			}
		})
	}
}

func freshIDCommand(t *testing.T, path string) *cobra.Command {
	t.Helper()
	original, err := resolveSchemaTarget(strings.Fields(path))
	if err != nil {
		t.Fatal(err)
	}
	root := &cobra.Command{Use: "hab"}
	current := root
	parts := strings.Fields(path)
	for i, part := range parts {
		child := &cobra.Command{Use: part}
		if i == len(parts)-1 {
			child.Use = original.Use
			child.Annotations = original.Annotations
		}
		current.AddCommand(child)
		current = child
	}
	original.Flags().VisitAll(func(f *pflag.Flag) {
		switch f.Value.Type() {
		case "string":
			current.Flags().String(f.Name, f.DefValue, f.Usage)
		case "stringArray":
			current.Flags().StringArray(f.Name, nil, f.Usage)
		case "stringSlice":
			current.Flags().StringSlice(f.Name, nil, f.Usage)
		case "bool":
			current.Flags().Bool(f.Name, false, f.Usage)
		}
	})
	return current
}
func TestResolveShortSelectors(t *testing.T) {
	full := "abcdef00111111111111111111111111"
	for _, tt := range []struct {
		path        string
		args, flags []string
		ns          idNamespace
	}{
		{"device get", []string{"abcdef00"}, nil, idNamespace{Kind: "device"}},
		{"device delete", []string{"abcdef00"}, []string{"--plan"}, idNamespace{Kind: "device"}},
		{"entity list", nil, []string{"--device", "abcdef00"}, idNamespace{Kind: "device"}},
		{"area update", []string{"kitchen"}, []string{"--labels", "abcdef00"}, idNamespace{Kind: "label"}},
		{"category assign", []string{"abcdef00", "automation.lights"}, nil, idNamespace{Kind: "category", Scope: "automation"}},
		{"todo remove", []string{"todo.shopping", "abcdef00"}, nil, idNamespace{Kind: "todo_item", Scope: "todo.shopping"}},
		{"search related", []string{"device", "abcdef00"}, nil, idNamespace{Kind: "device"}},
		{"search related", nil, []string{"--type", "device", "--id", "abcdef00"}, idNamespace{Kind: "device"}},
	} {
		t.Run(tt.path, func(t *testing.T) {
			cmd := freshIDCommand(t, tt.path)
			if err := cmd.ParseFlags(tt.flags); err != nil {
				t.Fatal(err)
			}
			args := append([]string{}, tt.args...)
			calls := 0
			err := resolveShortIDSelectors(cmd, args, func(ns idNamespace) ([]shortid.Candidate, error) {
				calls++
				if ns != tt.ns {
					t.Fatalf("namespace=%+v want=%+v", ns, tt.ns)
				}
				return []shortid.Candidate{{ID: full}}, nil
			})
			if err != nil || calls != 1 {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
			values := strings.Join(args, " ")
			cmd.Flags().Visit(func(f *pflag.Flag) { values += " " + f.Value.String() })
			if !strings.Contains(values, full) {
				t.Fatalf("ID was not expanded: %s", values)
			}
		})
	}
}
func TestShortIDsAmbiguityStopsCommand(t *testing.T) {
	cmd := freshIDCommand(t, "device delete")
	args := []string{"abcdef00"}
	err := resolveShortIDSelectors(cmd, args, func(idNamespace) ([]shortid.Candidate, error) {
		return []shortid.Candidate{{ID: "abcdef00111111111111111111111111", Name: "One"}, {ID: "abcdef00222222222222222222222222", Name: "Two"}}, nil
	})
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "ID_RESOLUTION_FAILED" {
		t.Fatalf("error=%v", err)
	}
	if args[0] != "abcdef00" {
		t.Fatal("ambiguous ID was changed")
	}
}
func TestCreationIDsAndReadableValuesArePreserved(t *testing.T) {
	for _, path := range []string{"automation create", "script create", "scene create", "entity update"} {
		cmd := freshIDCommand(t, path)
		if err := resolveShortIDSelectors(cmd, []string{"abcdef00"}, func(idNamespace) ([]shortid.Candidate, error) {
			t.Fatal("looked up a new ID or entity ID")
			return nil, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
}
func TestIDReferences(t *testing.T) {
	full := "abcdef00111111111111111111111111"
	data := map[string]any{"id": "abcdef00", "unique_id": "abcdef00", "name": "abcdef00", "actions": []any{map[string]any{"target": map[string]any{"device_id": []any{"abcdef00"}}}}, "variables": map[string]any{"device_id": "abcdef00"}}
	calls := 0
	err := walkIDReferences(data, func(ns idNamespace) ([]shortid.Candidate, error) {
		calls++
		if ns.Kind != "device" {
			t.Fatal(ns)
		}
		return []shortid.Candidate{{ID: full}}, nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
	target := data["actions"].([]any)[0].(map[string]any)["target"].(map[string]any)
	if !reflect.DeepEqual(target["device_id"], []any{full}) || data["id"] != "abcdef00" || data["name"] != "abcdef00" || data["variables"].(map[string]any)["device_id"] != "abcdef00" {
		t.Fatalf("wrong fields rewritten: %v", data)
	}
}
func TestIDCatalogCachesNamespaces(t *testing.T) {
	ws := &shortIDStore{}
	catalog := &idCatalog{ws: ws}
	for i := 0; i < 2; i++ {
		if _, err := catalog.list(idNamespace{Kind: "device"}); err != nil {
			t.Fatal(err)
		}
	}
	if ws.calls != 1 {
		t.Fatalf("discovery calls=%d", ws.calls)
	}
}

type shortIDStore struct {
	client.WebSocketAPI
	calls int
}

func (s *shortIDStore) DeviceRegistryList() ([]any, error) {
	s.calls++
	return []any{map[string]any{"id": "abcdef00111111111111111111111111", "name": "Kitchen"}}, nil
}

func TestCalendarPrefixWithWindow(t *testing.T) {
	// A 16-character hex prefix can itself look like a full opaque ID.
	full := "abcdef00111111112222222222222222"
	cmd := freshIDCommand(t, "calendar delete")
	if err := cmd.ParseFlags([]string{"--start", "2099-01-01", "--end", "2099-02-01"}); err != nil {
		t.Fatal(err)
	}
	args := []string{"calendar.personal", full[:16]}
	err := resolveShortIDSelectors(cmd, args, func(ns idNamespace) ([]shortid.Candidate, error) {
		if ns.Scope != "calendar.personal\n2099-01-01\n2099-02-01" {
			t.Fatal(ns)
		}
		return []shortid.Candidate{{ID: full}}, nil
	})
	if err != nil || args[1] != full {
		t.Fatalf("args=%v err=%v", args, err)
	}
}

func TestCategoryReferencesUseScope(t *testing.T) {
	full := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	data := map[string]any{"categories": map[string]any{"automation": full[:8]}}
	err := walkIDReferences(data, func(ns idNamespace) ([]shortid.Candidate, error) {
		if ns != (idNamespace{Kind: "category", Scope: "automation"}) {
			t.Fatal(ns)
		}
		return []shortid.Candidate{{ID: full}}, nil
	})
	if err != nil || data["categories"].(map[string]any)["automation"] != full {
		t.Fatalf("data=%v err=%v", data, err)
	}
}

type configIDStore struct{ client.WebSocketAPI }

func (s *configIDStore) EntityRegistryList() ([]any, error) {
	return []any{map[string]any{"entity_id": "automation.disabled", "platform": "automation", "unique_id": "abcdef00222222222222222222222222", "original_name": "Disabled automation"}}, nil
}
func TestDisabledConfigIDsParticipateInCollisions(t *testing.T) {
	states := []any{map[string]any{"entity_id": "automation.enabled", "attributes": map[string]any{"id": "abcdef00111111111111111111111111"}}}
	configs, err := allConfigIDEntries(&configIDStore{}, states, "automation")
	if err != nil {
		t.Fatal(err)
	}
	candidates := shortid.Candidates(configs, "id")
	if _, err := shortid.Resolve("abcdef00", candidates); err == nil {
		t.Fatal("ignored disabled automation when resolving prefix")
	}
	items := []map[string]any{{"entity_id": "automation.enabled"}}
	addConfigShortIDs(items, configs)
	if items[0]["short_id"] != "abcdef001" {
		t.Fatalf("unsafe discovery: %v", items)
	}
}

func TestEntityUpdateScopedCategoryPrefixes(t *testing.T) {
	cmd := freshIDCommand(t, "entity update")
	if err := cmd.ParseFlags([]string{"--category", "helpers=abcdef00", "--category", "script="}); err != nil {
		t.Fatal(err)
	}
	full := "abcdef00111111111111111111111111"
	calls := 0
	err := resolveShortIDSelectors(cmd, []string{"input_boolean.test"}, func(ns idNamespace) ([]shortid.Candidate, error) {
		calls++
		if ns != (idNamespace{Kind: "category", Scope: "helpers"}) {
			t.Fatal(ns)
		}
		return []shortid.Candidate{{ID: full}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	values, err := cmd.Flags().GetStringArray("category")
	if err != nil || calls != 1 || !reflect.DeepEqual(values, []string{"helpers=" + full, "script="}) {
		t.Fatalf("values=%v calls=%d err=%v", values, calls, err)
	}
}

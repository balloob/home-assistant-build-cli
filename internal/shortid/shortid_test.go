package shortid

import (
	"errors"
	"reflect"
	"testing"
)

func TestResolve(t *testing.T) {
	candidates := []Candidate{
		{ID: "abcdef00111111111111111111111111", Name: "Kitchen"},
		{ID: "abcdef00222222222222222222222222", Name: "Bedroom"},
		{ID: "12345678", Name: "Exact short ID"},
		{ID: "12345678999999999999999999999999", Name: "Long ID"},
		{ID: "living_room", Name: "Readable ID"},
	}
	for _, tt := range []struct {
		input, want string
		matches     int
	}{
		{"abcdef001", "abcdef00111111111111111111111111", 0},
		{"abcdef00111111111111111111111111", "abcdef00111111111111111111111111", 0},
		{"abcdef00", "", 2}, {"abcdef", "", 0}, {"ffffffff", "", 0},
		{"12345678", "12345678", 0}, {"living_room", "living_room", 0}, {"living_r", "", 0},
	} {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Resolve(tt.input, candidates)
			if tt.want != "" {
				if err != nil || got != tt.want {
					t.Fatalf("got=%q err=%v", got, err)
				}
				return
			}
			var resolution *Error
			if !errors.As(err, &resolution) || len(resolution.Candidates) != tt.matches {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
func TestAbbreviations(t *testing.T) {
	candidates := []Candidate{
		{ID: "01K8EXAMPLE0000000000000001"}, {ID: "01K8EXAMPLE0000000000000002"}, // L is not valid in a ULID
		{ID: "12345678"}, {ID: "12345678999999999999999999999999"},
		{ID: "aabbccdd-1234-4321-abcd-111111111111"},
		{ID: "living_room"},
	}
	shorts := Abbreviations(candidates)
	if shorts["12345678999999999999999999999999"] != "123456789" {
		t.Fatalf("exact ID collision: %v", shorts)
	}
	if shorts["aabbccdd-1234-4321-abcd-111111111111"] != "aabbccdd" {
		t.Fatalf("UUID: %v", shorts)
	}
	if _, ok := shorts["living_room"]; ok {
		t.Fatal("abbreviated a readable ID")
	}
	for id, short := range shorts {
		full, err := Resolve(short, candidates)
		if err != nil || full != id {
			t.Fatalf("%s -> %s: %v", id, short, err)
		}
	}
}
func TestULIDSharedTimestamp(t *testing.T) {
	ids := []Candidate{{ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV"}, {ID: "01ARZ3NDEKTSV4RRFFQ69G5FAW"}}
	shorts := Abbreviations(ids)
	if len(shorts) != 2 || shorts[ids[0].ID] != ids[0].ID || shorts[ids[1].ID] != ids[1].ID {
		t.Fatalf("unsafe ULID abbreviations: %v", shorts)
	}
}
func TestAnnotateBeforeLimit(t *testing.T) {
	entries := []any{map[string]any{"id": "abcdef00111111111111111111111111"}, map[string]any{"id": "abcdef00222222222222222222222222"}}
	annotated := Annotate(entries, "id")[:1]
	if annotated[0].(map[string]any)["short_id"] != "abcdef001" {
		t.Fatalf("bad abbreviation: %v", annotated)
	}
	if _, ok := entries[0].(map[string]any)["short_id"]; ok {
		t.Fatal("mutated API response")
	}
}
func TestDuplicateIDs(t *testing.T) {
	id := "abcdef00111111111111111111111111"
	got, err := Resolve("abcdef00", []Candidate{{ID: id}, {ID: id}})
	if err != nil || got != id {
		t.Fatalf("duplicate entry: %s %v", got, err)
	}
}
func TestPossible(t *testing.T) {
	for _, id := range []string{"light.kitchen", "living_room", "aabb", "my-long-readable-name", "Kitchen"} {
		if Possible(id) {
			t.Errorf("not opaque: %s", id)
		}
	}
	if !reflect.DeepEqual(Annotate(nil, "id"), []any{}) {
		t.Fatal("empty discovery must stay an array")
	}
}

func TestNamespacedOpaqueIDs(t *testing.T) {
	ids := []Candidate{{ID: "sftp_storage.01ARZ3NDEKTSV4RRFFQ69G5FAV"}, {ID: "backup.local"}, {ID: "cloud.cloudflare"}}
	shorts := Abbreviations(ids)
	if len(shorts) != 1 || shorts[ids[0].ID] != "sftp_storage.01ARZ3ND" {
		t.Fatalf("backup agent abbreviations: %v", shorts)
	}
	if !Possible(shorts[ids[0].ID]) {
		t.Fatal("generated abbreviation is not recognized")
	}
	got, err := Resolve(shorts[ids[0].ID], ids)
	if err != nil || got != ids[0].ID {
		t.Fatalf("resolution=%s err=%v", got, err)
	}
}

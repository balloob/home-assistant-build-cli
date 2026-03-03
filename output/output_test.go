package output

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"entity_id", "Entity Id"},
		{"name", "Name"},
		{"friendly_name", "Friendly Name"},
		{"", ""},
		{"single", "Single"},
		{"a_b_c", "A B C"},
	}
	for _, tt := range tests {
		got := formatKey(tt.input)
		if got != tt.want {
			t.Errorf("formatKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"bool true", true, "Yes"},
		{"bool false", false, "No"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"map", map[string]interface{}{"a": 1}, "..."},
		{"slice", []interface{}{1, 2}, "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.input)
			if got != tt.want {
				t.Errorf("formatValue(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  string
	}{
		{"friendly_name", map[string]interface{}{"friendly_name": "My Lamp", "name": "lamp"}, "My Lamp"},
		{"name", map[string]interface{}{"name": "lamp", "entity_id": "light.lamp"}, "lamp"},
		{"entity_id", map[string]interface{}{"entity_id": "light.lamp", "id": "abc"}, "light.lamp"},
		{"id", map[string]interface{}{"id": "abc"}, "abc"},
		{"fallback", map[string]interface{}{"foo": "bar"}, "map[foo:bar]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDisplayName(tt.input)
			if got != tt.want {
				t.Errorf("getDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatText(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		message string
		want    string
	}{
		{"message takes priority", "ignored", "hello", "hello"},
		{"nil data", nil, "", "Done."},
		{"string data", "hello world", "", "hello world"},
		{"bool true", true, "", "Yes"},
		{"bool false", false, "", "No"},
		{"int", 42, "", "42"},
		{"float", 3.14, "", "3.14"},
		{"empty list", []interface{}{}, "", "No items."},
		{"simple list", []interface{}{"a", "b"}, "", "  - a\n  - b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatText(tt.data, tt.message)
			if got != tt.want {
				t.Errorf("formatText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDict_DeterministicOrder(t *testing.T) {
	data := map[string]interface{}{
		"zebra":  "z",
		"apple":  "a",
		"mango":  "m",
		"banana": "b",
	}

	// Run multiple times to verify deterministic output
	first := formatDict(data)
	for i := 0; i < 20; i++ {
		got := formatDict(data)
		if got != first {
			t.Errorf("formatDict produced non-deterministic output on iteration %d:\nfirst: %q\ngot:   %q", i, first, got)
		}
	}

	// Verify sorted order
	lines := strings.Split(first, "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d: %q", len(lines), first)
	}
	if !strings.HasPrefix(lines[0], "Apple:") {
		t.Errorf("first line should start with 'Apple:', got %q", lines[0])
	}
	if !strings.HasPrefix(lines[3], "Zebra:") {
		t.Errorf("last line should start with 'Zebra:', got %q", lines[3])
	}
}

func TestFormatDict_NestedMapSorted(t *testing.T) {
	data := map[string]interface{}{
		"info": map[string]interface{}{
			"z_field": "z",
			"a_field": "a",
		},
	}

	result := formatDict(data)
	lines := strings.Split(result, "\n")

	// First line: "Info:"
	if lines[0] != "Info:" {
		t.Errorf("expected 'Info:', got %q", lines[0])
	}
	// Nested keys should be sorted: a_field before z_field
	if !strings.Contains(lines[1], "a_field") {
		t.Errorf("expected a_field first, got %q", lines[1])
	}
	if !strings.Contains(lines[2], "z_field") {
		t.Errorf("expected z_field second, got %q", lines[2])
	}
}

func TestFormatDictList_DeterministicColumns(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{
			"name":      "Alice",
			"entity_id": "person.alice",
			"age":       float64(30),
		},
		map[string]interface{}{
			"name":      "Bob",
			"entity_id": "person.bob",
			"age":       float64(25),
		},
	}

	first := formatDictList(data)
	for i := 0; i < 20; i++ {
		got := formatDictList(data)
		if got != first {
			t.Errorf("formatDictList produced non-deterministic output on iteration %d", i)
		}
	}

	// Verify columns are alphabetically sorted
	lines := strings.Split(first, "\n")
	header := lines[0]
	// Should be: "Age | Entity Id | Name" (sorted)
	if !strings.HasPrefix(header, "Age") {
		t.Errorf("expected header to start with 'Age', got %q", header)
	}
}

func TestFormatDictList_Empty(t *testing.T) {
	got := formatDictList([]interface{}{})
	if got != "No items." {
		t.Errorf("formatDictList([]) = %q, want %q", got, "No items.")
	}
}

func TestFormatDictList_LimitColumns(t *testing.T) {
	// Create a map with 8 keys — should be limited to 6
	item := map[string]interface{}{
		"a": "1", "b": "2", "c": "3", "d": "4",
		"e": "5", "f": "6", "g": "7", "h": "8",
	}
	data := []interface{}{item}

	result := formatDictList(data)
	header := strings.Split(result, "\n")[0]
	// Count columns by counting separators
	cols := strings.Count(header, " | ") + 1
	if cols != 6 {
		t.Errorf("expected 6 columns, got %d (header: %q)", cols, header)
	}
}

func TestFormatOutput_TextMode(t *testing.T) {
	got := FormatOutput("hello", true, "")
	if got != "hello" {
		t.Errorf("FormatOutput text mode = %q, want %q", got, "hello")
	}
}

func TestFormatOutput_JSONMode(t *testing.T) {
	got := FormatOutput("hello", false, "test msg")

	var resp Response
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Message != "test msg" {
		t.Errorf("expected message 'test msg', got %q", resp.Message)
	}
	if resp.Metadata == nil || resp.Metadata["timestamp"] == nil {
		t.Error("expected metadata with timestamp")
	}
}

func TestFormatError(t *testing.T) {
	got := FormatError("NOT_FOUND", "entity not found", nil)

	var resp Response
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error == nil {
		t.Fatal("expected error detail")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %q", resp.Error.Code)
	}
	if resp.Error.Message != "entity not found" {
		t.Errorf("expected message 'entity not found', got %q", resp.Error.Message)
	}
}

func TestFormatErrorText(t *testing.T) {
	got := FormatErrorText("something broke", "try again")
	if got != "Error: something broke\nSuggestion: try again" {
		t.Errorf("unexpected error text: %q", got)
	}

	got2 := FormatErrorText("something broke", "")
	if got2 != "Error: something broke" {
		t.Errorf("unexpected error text without suggestion: %q", got2)
	}
}

func TestFormatSuccess(t *testing.T) {
	got := FormatSuccess(map[string]interface{}{"id": "abc"}, "created")

	var resp Response
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Message != "created" {
		t.Errorf("expected message 'created', got %q", resp.Message)
	}
}

// Additional coverage tests

func TestFormatText_Int64(t *testing.T) {
	got := formatText(int64(999), "")
	if got != "999" {
		t.Errorf("formatText(int64) = %q, want %q", got, "999")
	}
}

func TestFormatText_MapData(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"id":   "abc",
	}
	got := formatText(data, "")
	// Should call formatDict — verify keys appear
	if !strings.Contains(got, "Id:") || !strings.Contains(got, "Name:") {
		t.Errorf("formatText(map) should format as dict, got %q", got)
	}
}

func TestFormatText_CustomStructViaMarshal(t *testing.T) {
	// The default branch in formatText marshals to JSON and re-parses
	type custom struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	got := formatText(custom{Name: "alice", Age: 30}, "")
	// Should contain formatted output from the re-parsed map
	if got == "" {
		t.Error("expected non-empty output for custom struct")
	}
}

func TestFormatList_MixedItems(t *testing.T) {
	// A list where not all items are maps — triggers the simple list path
	data := []interface{}{"hello", 42, true}
	got := formatList(data)
	if !strings.Contains(got, "- hello") {
		t.Errorf("expected '- hello' in output, got %q", got)
	}
	if !strings.Contains(got, "- 42") {
		t.Errorf("expected '- 42' in output, got %q", got)
	}
}

func TestFormatList_MapItemsInSimpleList(t *testing.T) {
	// Simple list with a map item that isn't allMaps
	data := []interface{}{
		map[string]interface{}{"name": "lamp"},
		"plain string",
	}
	got := formatList(data)
	if !strings.Contains(got, "- lamp") {
		t.Errorf("expected map display name 'lamp', got %q", got)
	}
	if !strings.Contains(got, "- plain string") {
		t.Errorf("expected plain string item, got %q", got)
	}
}

func TestFormatDictList_MaxRows(t *testing.T) {
	// Create >50 items to trigger the "and N more" truncation
	var data []interface{}
	for i := 0; i < 55; i++ {
		data = append(data, map[string]interface{}{"id": i})
	}
	got := formatDictList(data)
	if !strings.Contains(got, "... and 5 more items") {
		t.Errorf("expected truncation message, got %q", got)
	}
}

func TestFormatDictList_LongValues(t *testing.T) {
	// Values >30 chars should be truncated
	data := []interface{}{
		map[string]interface{}{
			"desc": "This is a very long description that exceeds thirty characters",
		},
	}
	got := formatDictList(data)
	if !strings.Contains(got, "...") {
		t.Errorf("expected truncated value, got %q", got)
	}
}

func TestFormatDictList_NonMapFirstItem(t *testing.T) {
	data := []interface{}{"not a map"}
	got := formatDictList(data)
	if got != "No items." {
		t.Errorf("expected 'No items.' for non-map first item, got %q", got)
	}
}

func TestFormatDict_ArrayValue(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}
	got := formatDict(data)
	if !strings.Contains(got, "Items:") {
		t.Errorf("expected 'Items:' header, got %q", got)
	}
	if !strings.Contains(got, "- a") {
		t.Errorf("expected list item '- a', got %q", got)
	}
}

func TestFormatDict_ArrayWithMapItems(t *testing.T) {
	data := map[string]interface{}{
		"devices": []interface{}{
			map[string]interface{}{"name": "Lamp"},
			map[string]interface{}{"name": "Switch"},
		},
	}
	got := formatDict(data)
	if !strings.Contains(got, "- Lamp") {
		t.Errorf("expected '- Lamp', got %q", got)
	}
	if !strings.Contains(got, "- Switch") {
		t.Errorf("expected '- Switch', got %q", got)
	}
}

func TestFormatDict_ArrayTruncation(t *testing.T) {
	// Array with >10 items should be truncated
	var items []interface{}
	for i := 0; i < 15; i++ {
		items = append(items, i)
	}
	data := map[string]interface{}{
		"numbers": items,
	}
	got := formatDict(data)
	if !strings.Contains(got, "... and 5 more") {
		t.Errorf("expected truncation message, got %q", got)
	}
}

func TestFormatError_WithDetails(t *testing.T) {
	details := map[string]interface{}{"field": "name", "reason": "required"}
	got := FormatError("VALIDATION_ERROR", "invalid input", details)

	var resp Response
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error detail")
	}
	if resp.Error.Details == nil {
		t.Fatal("expected error details")
	}
	if resp.Error.Details["field"] != "name" {
		t.Errorf("expected field=name, got %v", resp.Error.Details["field"])
	}
}

func TestFormatOutput_JSONWithNilData(t *testing.T) {
	got := FormatOutput(nil, false, "operation complete")
	var resp Response
	if err := json.Unmarshal([]byte(got), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Message != "operation complete" {
		t.Errorf("expected message, got %q", resp.Message)
	}
}

func TestFormatOutput_TextWithMessage(t *testing.T) {
	got := FormatOutput(nil, true, "All done")
	if got != "All done" {
		t.Errorf("FormatOutput text with message = %q, want %q", got, "All done")
	}
}

func TestFormatOutput_TextWithNilNoMessage(t *testing.T) {
	got := FormatOutput(nil, true, "")
	if got != "Done." {
		t.Errorf("FormatOutput text nil = %q, want %q", got, "Done.")
	}
}

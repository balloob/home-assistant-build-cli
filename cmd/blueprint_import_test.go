package cmd

import (
	"strings"
	"testing"
)

type fakeBlueprintWS struct {
	importResult map[string]interface{}
	calls        []string
	saveParams   map[string]interface{}
}

func (f *fakeBlueprintWS) SendCommand(cmdType string, params map[string]interface{}) (interface{}, error) {
	f.calls = append(f.calls, cmdType)
	if cmdType == "blueprint/save" {
		f.saveParams = params
		return map[string]interface{}{"overrides_existing": params["allow_override"]}, nil
	}
	return f.importResult, nil
}

func newFakeBlueprintImport(exists bool, validationErrors interface{}) *fakeBlueprintWS {
	return &fakeBlueprintWS{importResult: map[string]interface{}{
		"suggested_filename": "home-assistant/motion_light",
		"raw_data":           "blueprint:\n  name: Motion\n  domain: automation\n",
		"blueprint": map[string]interface{}{
			"metadata": map[string]interface{}{"domain": "automation"},
		},
		"validation_errors": validationErrors,
		"exists":            exists,
	}}
}

const testBlueprintURL = "https://example.com/motion_light.yaml"

func TestImportBlueprintSaves(t *testing.T) {
	ws := newFakeBlueprintImport(false, nil)
	result, err := importBlueprint(ws, testBlueprintURL, false)
	if err != nil {
		t.Fatalf("importBlueprint: %v", err)
	}
	if strings.Join(ws.calls, ",") != "blueprint/import,blueprint/save" {
		t.Fatalf("calls = %v", ws.calls)
	}
	want := map[string]interface{}{
		"domain":         "automation",
		"path":           "home-assistant/motion_light.yaml",
		"yaml":           ws.importResult["raw_data"],
		"source_url":     testBlueprintURL,
		"allow_override": false,
	}
	for k, v := range want {
		if ws.saveParams[k] != v {
			t.Errorf("save %s = %v, want %v", k, ws.saveParams[k], v)
		}
	}
	if result["path"] != "home-assistant/motion_light.yaml" || result["domain"] != "automation" {
		t.Errorf("result = %v", result)
	}
}

func TestImportBlueprintValidationErrors(t *testing.T) {
	ws := newFakeBlueprintImport(false, []interface{}{"Requires at least Home Assistant 2099.1.0"})
	_, err := importBlueprint(ws, testBlueprintURL, false)
	if err == nil || !strings.Contains(err.Error(), "Requires at least Home Assistant 2099.1.0") {
		t.Fatalf("err = %v", err)
	}
	if len(ws.calls) != 1 {
		t.Errorf("save must not be sent, calls = %v", ws.calls)
	}
}

func TestImportBlueprintExistsNeedsOverride(t *testing.T) {
	ws := newFakeBlueprintImport(true, nil)
	_, err := importBlueprint(ws, testBlueprintURL, false)
	if err == nil || !strings.Contains(err.Error(), "--override") {
		t.Fatalf("err = %v", err)
	}
	if len(ws.calls) != 1 {
		t.Errorf("save must not be sent, calls = %v", ws.calls)
	}

	ws = newFakeBlueprintImport(true, nil)
	result, err := importBlueprint(ws, testBlueprintURL, true)
	if err != nil {
		t.Fatalf("importBlueprint with override: %v", err)
	}
	if ws.saveParams["allow_override"] != true || result["overrides_existing"] != true {
		t.Errorf("save = %v, result = %v", ws.saveParams, result)
	}
}

package esphometasmota

import "testing"

func TestParseTemplate(t *testing.T) {
	tpl, err := ParseTemplate(`{"NAME":"Example","GPIO":[416,0,418,0,417,2720,0,0,2624,32,2656,224,0,0],"FLAG":0,"BASE":45}`)
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	if tpl.Name != "Example" {
		t.Fatalf("Name = %q, want Example", tpl.Name)
	}
	if len(tpl.GPIO) != 14 {
		t.Fatalf("len(GPIO) = %d, want 14", len(tpl.GPIO))
	}
}

func TestAnalyze(t *testing.T) {
	tpl, err := ParseTemplate(`{"NAME":"Example","GPIO":[0,0,0,0,0,0,0,0,0,32,0,224,0,0],"FLAG":0,"BASE":18}`)
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}

	analysis := Analyze(tpl)
	if analysis.SupportedCount != 2 {
		t.Fatalf("SupportedCount = %d, want 2", analysis.SupportedCount)
	}
	if analysis.UnsupportedCount != 0 {
		t.Fatalf("UnsupportedCount = %d, want 0", analysis.UnsupportedCount)
	}
}

func TestBuildOverlay(t *testing.T) {
	tpl, err := ParseTemplate(`{"NAME":"Example","GPIO":[0,0,0,0,0,0,0,0,0,32,0,224,0,0],"FLAG":0,"BASE":18}`)
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}

	overlay, analysis := BuildOverlay(tpl)
	if analysis.SupportedCount != 2 {
		t.Fatalf("SupportedCount = %d, want 2", analysis.SupportedCount)
	}

	switches, ok := overlay["switch"].([]any)
	if !ok || len(switches) != 1 {
		t.Fatalf("overlay switch = %#v", overlay["switch"])
	}

	binaries, ok := overlay["binary_sensor"].([]any)
	if !ok || len(binaries) != 1 {
		t.Fatalf("overlay binary_sensor = %#v", overlay["binary_sensor"])
	}
}

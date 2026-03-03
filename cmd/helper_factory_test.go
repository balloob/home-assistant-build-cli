package cmd

import "testing"

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		hours int
		mins  int
		secs  int
	}{
		{"1:30:00", 1, 30, 0},
		{"0:05:30", 0, 5, 30},
		{"0:00:00", 0, 0, 0},
		{"24:00:00", 24, 0, 0},
		{"0:0:0", 0, 0, 0},
		{"12:59:59", 12, 59, 59},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseDuration(tt.input)
			if got["hours"] != tt.hours {
				t.Errorf("hours = %v, want %v", got["hours"], tt.hours)
			}
			if got["minutes"] != tt.mins {
				t.Errorf("minutes = %v, want %v", got["minutes"], tt.mins)
			}
			if got["seconds"] != tt.secs {
				t.Errorf("seconds = %v, want %v", got["seconds"], tt.secs)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"", ""},
		{"a", "A"},
		{"Hello", "Hello"},
		{"input boolean", "Input boolean"},
		{"UPPER", "UPPER"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := capitalize(tt.input)
			if got != tt.want {
				t.Errorf("capitalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

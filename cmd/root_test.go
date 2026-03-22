package cmd

import "testing"

func TestExecutableName(t *testing.T) {
	tests := []struct {
		name string
		arg0 string
		want string
	}{
		{name: "empty arg", arg0: "", want: "hab"},
		{name: "unix path", arg0: "/usr/local/bin/hab", want: "hab"},
		{name: "windows path", arg0: `C:\Program Files\hab\hab.exe`, want: "hab.exe"},
		{name: "bare name", arg0: "hab", want: "hab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := executableName(tt.arg0); got != tt.want {
				t.Errorf("executableName(%q) = %q, want %q", tt.arg0, got, tt.want)
			}
		})
	}
}

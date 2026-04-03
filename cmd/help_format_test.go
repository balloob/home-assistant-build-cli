package cmd

import "testing"

func TestFormatExamples(t *testing.T) {
	tests := []struct {
		name     string
		examples string
		want     string
	}{
		{
			name: "normalizes common indent",
			examples: `  hab action data
  hab action data weather
  hab action data --json`,
			want: `  hab action data
  hab action data weather
  hab action data --json`,
		},
		{
			name: "adds indent when none exists",
			examples: `hab guide
hab guide list --json`,
			want: `  hab guide
  hab guide list --json`,
		},
		{
			name: "preserves relative continuation indent",
			examples: `hab automation create-from-blueprint my_motion_light homeassistant/motion_light \
   -d '{"alias":"Kitchen Motion Light"}'`,
			want: `  hab automation create-from-blueprint my_motion_light homeassistant/motion_light \
     -d '{"alias":"Kitchen Motion Light"}'`,
		},
		{
			name:     "trims surrounding blank lines",
			examples: "\n\n  hab overview\n\n",
			want:     "  hab overview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatExamples(tt.examples)
			if got != tt.want {
				t.Fatalf("formatExamples() = %q, want %q", got, tt.want)
			}
		})
	}
}

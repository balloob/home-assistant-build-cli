package cmd

import (
	"testing"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestDetermineOutputMode(t *testing.T) {
	origStdout := stdoutIsTerminalFunc
	defer func() { stdoutIsTerminalFunc = origStdout }()

	tests := []struct {
		name         string
		interactive  bool
		args         []string
		wantText     bool
		wantJSON     bool
		wantModeName string
	}{
		{name: "interactive default text", interactive: true, wantText: true, wantJSON: false, wantModeName: "text"},
		{name: "noninteractive default json", interactive: false, wantText: false, wantJSON: true, wantModeName: "json"},
		{name: "noninteractive explicit text", interactive: false, args: []string{"--text"}, wantText: true, wantJSON: false, wantModeName: "text"},
		{name: "interactive explicit json", interactive: true, args: []string{"--json"}, wantText: false, wantJSON: true, wantModeName: "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdoutIsTerminalFunc = func() bool { return tt.interactive }
			viper.Set("json", false)
			viper.Set("text", false)

			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().Bool("json", false, "")
			cmd.Flags().Bool("text", false, "")
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("parse flags: %v", err)
			}
			jsonFlag, _ := cmd.Flags().GetBool("json")
			textFlag, _ := cmd.Flags().GetBool("text")
			viper.Set("json", jsonFlag)
			viper.Set("text", textFlag)

			gotText, gotJSON, gotMode, err := determineOutputMode(cmd)
			if err != nil {
				t.Fatalf("determineOutputMode error: %v", err)
			}
			if gotText != tt.wantText || gotJSON != tt.wantJSON || gotMode != tt.wantModeName {
				t.Fatalf("got (text=%v json=%v mode=%q), want (text=%v json=%v mode=%q)", gotText, gotJSON, gotMode, tt.wantText, tt.wantJSON, tt.wantModeName)
			}
		})
	}
}

func TestResolveArgConflictsOnMismatch(t *testing.T) {
	_, err := resolveArg("flag-value", []string{"positional-value"}, 0, "entity ID")
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestConfirmActionNonInteractiveRequiresForce(t *testing.T) {
	origStdin := stdinIsTerminalFunc
	defer func() { stdinIsTerminalFunc = origStdin }()
	stdinIsTerminalFunc = func() bool { return false }

	err := confirmAction(false, "Delete device abc?", "delete device")
	if err == nil {
		t.Fatal("expected confirmation required error")
	}
	apiErr, ok := err.(*client.APIError)
	if !ok || apiErr.Code != client.ErrCodeConfirmationRequired {
		t.Fatalf("got %#v, want confirmation required APIError", err)
	}
}

func TestConfirmActionInteractiveAcceptsAndCancels(t *testing.T) {
	origStdin := stdinIsTerminalFunc
	origReader := readConfirmationLine
	defer func() {
		stdinIsTerminalFunc = origStdin
		readConfirmationLine = origReader
	}()
	stdinIsTerminalFunc = func() bool { return true }

	readConfirmationLine = func() (string, error) { return "yes\n", nil }
	if err := confirmAction(false, "Delete area kitchen?", "delete area"); err != nil {
		t.Fatalf("expected confirmation success, got %v", err)
	}

	readConfirmationLine = func() (string, error) { return "no\n", nil }
	err := confirmAction(false, "Delete area kitchen?", "delete area")
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	apiErr, ok := err.(*client.APIError)
	if !ok || apiErr.Code != client.ErrCodeCancelled {
		t.Fatalf("got %#v, want cancelled APIError", err)
	}
}

func TestExecutionMetadataIncludesModeAuthAndFallback(t *testing.T) {
	resetExecutionMetadata()
	noteOutputMode("json")
	noteAuthSource("env_token")
	noteFallback("returned state only")

	metadata := getExecutionMetadata()
	if metadata["output_mode"] != "json" {
		t.Fatalf("metadata = %#v, want output_mode=json", metadata)
	}
	if metadata["auth_source"] != "env_token" {
		t.Fatalf("metadata = %#v, want auth_source=env_token", metadata)
	}
	if metadata["partial_result"] != true {
		t.Fatalf("metadata = %#v, want partial_result=true", metadata)
	}
}

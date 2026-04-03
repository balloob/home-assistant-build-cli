package cmd

import (
	"strings"
	"testing"

	"github.com/home-assistant/hab/guide"
)

func TestGuideSuggestedCommandsResolveToKnownCommands(t *testing.T) {
	for _, topic := range guide.List() {
		for _, command := range topic.SuggestedCommands {
			requireResolvableHabCommand(t, topic.ID, command)
		}
	}
}

func TestGuideCodeFenceCommandsResolveToKnownCommands(t *testing.T) {
	for _, topic := range guide.List() {
		content, err := guide.GetTopic(topic.ID)
		if err != nil {
			t.Fatalf("GetTopic(%q) error = %v", topic.ID, err)
		}

		for _, command := range extractHabCommandsFromCodeFences(content.Content) {
			requireResolvableHabCommand(t, topic.ID, command)
		}
	}
}

func requireResolvableHabCommand(t *testing.T, topicID string, command string) {
	t.Helper()

	fields := strings.Fields(command)
	if len(fields) < 2 {
		t.Fatalf("topic %q contains malformed command %q", topicID, command)
	}
	if fields[0] != "hab" {
		t.Fatalf("topic %q contains non-hab command %q", topicID, command)
	}

	resolvedCmd, _, err := rootCmd.Find(fields[1:])
	if err != nil {
		t.Fatalf("topic %q command %q failed to resolve: %v", topicID, command, err)
	}
	if resolvedCmd == nil || resolvedCmd == rootCmd {
		t.Fatalf("topic %q command %q did not resolve to a subcommand", topicID, command)
	}
}

func extractHabCommandsFromCodeFences(markdown string) []string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	commands := make([]string, 0)
	inFence := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if !inFence {
			continue
		}
		if strings.HasPrefix(trimmed, "hab ") {
			commands = append(commands, trimmed)
		}
	}

	return commands
}

package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/home-assistant/hab/guide"
	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var guideCmd = &cobra.Command{
	Use:     "guide [topic]",
	Short:   "Display built-in usage guides",
	Long:    "Display embedded guides for using hab effectively, including LLM-focused workflows.",
	Example: "hab guide\nhab guide discovery\nhab guide dashboard --json\nhab guide list --json",
	Args:    cobra.MaximumNArgs(1),
	GroupID: "start",
	RunE:    runGuide,
}

var guideListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available guide topics",
	Long:  "List all available guide topics and their summaries.",
	Args:  cobra.NoArgs,
	RunE:  runGuideList,
}

func init() {
	rootCmd.AddCommand(guideCmd)
	guideCmd.AddCommand(guideListCmd)
}

func runGuide(cmd *cobra.Command, args []string) error {
	topic := "index"
	if len(args) > 0 {
		topic = args[0]
	}

	return runGuideTopic(topic)
}

func runGuideList(cmd *cobra.Command, args []string) error {
	topics := guide.List()
	if getTextMode() {
		fmt.Println("Available guide topics:")
		for _, topic := range topics {
			line := fmt.Sprintf("- %s: %s", topic.ID, topic.Summary)
			if len(topic.Aliases) > 0 {
				line += fmt.Sprintf(" (aliases: %s)", strings.Join(topic.Aliases, ", "))
			}
			fmt.Println(line)
		}
		return nil
	}

	output.PrintOutput(buildGuideListOutput(topics), false, "")
	return nil
}

func buildGuideListOutput(topics []guide.Topic) []map[string]interface{} {
	result := make([]map[string]interface{}, len(topics))
	for i, topic := range topics {
		result[i] = map[string]interface{}{
			"id":                 topic.ID,
			"title":              topic.Title,
			"summary":            topic.Summary,
			"aliases":            topic.Aliases,
			"related_topics":     topic.RelatedTopics,
			"suggested_commands": topic.SuggestedCommands,
			"recipes":            topic.Recipes,
		}
	}

	return result
}

func runGuideTopic(topicName string) error {
	topic, err := guide.GetTopic(topicName)
	if err != nil {
		if errors.Is(err, guide.ErrTopicNotFound) {
			return fmt.Errorf("unknown guide topic '%s'. Available topics: %s", topicName, strings.Join(guide.TopicIDs(), ", "))
		}
		return fmt.Errorf("failed to load guide topic '%s': %w", topicName, err)
	}

	if getTextMode() {
		fmt.Print(topic.Content)
		return nil
	}

	output.PrintOutput(buildGuideTopicOutput(topic), false, "")
	return nil
}

func buildGuideTopicOutput(topic guide.TopicContent) map[string]interface{} {
	return map[string]interface{}{
		"topic":                 topic.Topic.ID,
		"title":                 topic.Topic.Title,
		"summary":               topic.Topic.Summary,
		"aliases":               topic.Topic.Aliases,
		"related_topics":        topic.Topic.RelatedTopics,
		"suggested_commands":    topic.Topic.SuggestedCommands,
		"prerequisites":         topic.Topic.Prerequisites,
		"discovery_steps":       topic.Topic.DiscoverySteps,
		"mutation_patterns":     topic.Topic.MutationPatterns,
		"verification_commands": topic.Topic.VerificationCommands,
		"pitfalls":              topic.Topic.Pitfalls,
		"recipes":               topic.Topic.Recipes,
		"content":               topic.Content,
	}
}

package cmd

import (
	"fmt"

	"github.com/home-assistant/hab/output"
	"github.com/spf13/cobra"
)

var marketplaceGithubConnectCmd = &cobra.Command{
	Use:   "github-connect",
	Short: "Connect a GitHub account to the Marketplace",
	Long: `Start connecting a GitHub account to the Marketplace.

The command returns a URL and a code. The user opens the URL, signs in to
GitHub, and enters the code. Home Assistant then stores the token by itself.
Run 'hab marketplace info' to check that github_connected is true.

A GitHub account is needed to add custom repositories, and it raises the
GitHub rate limit.`,
	RunE: runMarketplaceGithubConnect,
}

func init() {
	marketplaceCmd.AddCommand(marketplaceGithubConnectCmd)
	mergeSchemaAnnotation(marketplaceGithubConnectCmd, SchemaAnnotation{
		SideEffect:   "write",
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws", "rest"},
		ResourceType: "marketplace",
		InputSources: []string{"flags"},
	})
}

func runMarketplaceGithubConnect(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	ws, err := getWSClient()
	if err != nil {
		return err
	}
	defer ws.Close()

	rest, err := getRESTClient()
	if err != nil {
		return err
	}

	result, err := ws.SendCommand("marketplace/github/connect", nil)
	if err != nil {
		return err
	}
	started, _ := result.(map[string]interface{})
	flowID, _ := started["flow_id"].(string)
	if flowID == "" {
		return fmt.Errorf("marketplace/github/connect returned no flow_id")
	}

	// The flow opens on an empty confirmation form. Submitting it asks GitHub
	// for a device code.
	step, err := rest.ConfigFlowStep(flowID, map[string]interface{}{})
	if err != nil {
		return err
	}
	placeholders, _ := step["description_placeholders"].(map[string]interface{})
	url, _ := placeholders["url"].(string)
	code, _ := placeholders["code"].(string)
	if url == "" || code == "" {
		return fmt.Errorf("GitHub did not return a device code: %v", step)
	}

	output.PrintSuccessWithContext(map[string]interface{}{
		"flow_id": flowID,
		"url":     url,
		"code":    code,
	}, textMode, fmt.Sprintf("Open %s and enter the code %s.", url, code), output.EnvelopeContext{
		Operation:            "github-connect",
		ResourceType:         "marketplace",
		VerificationCommands: []string{"hab marketplace info --json"},
	})
	return nil
}

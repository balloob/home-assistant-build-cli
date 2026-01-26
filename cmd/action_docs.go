package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var actionDocsName string

var actionDocsCmd = &cobra.Command{
	Use:   "docs [domain.action]",
	Short: "Show action documentation",
	Long:  `Show the documentation for a specific action including available fields.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runActionDocs,
}

func init() {
	actionCmd.AddCommand(actionDocsCmd)
	actionDocsCmd.Flags().StringVar(&actionDocsName, "action", "", "Action name in domain.action format")
}

func runActionDocs(cmd *cobra.Command, args []string) error {
	actionName := actionDocsName
	if actionName == "" && len(args) > 0 {
		actionName = args[0]
	}
	if actionName == "" {
		return fmt.Errorf("action name is required (use --action flag or positional argument)")
	}
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")

	// Parse action name
	parts := strings.SplitN(actionName, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid action format: %s. Expected domain.action", actionName)
	}
	domain := parts[0]
	service := parts[1]

	manager := auth.NewManager(configDir)
	restClient, err := manager.GetRestClient()
	if err != nil {
		return err
	}

	services, err := restClient.GetServices()
	if err != nil {
		return err
	}

	// Find the specific service
	for _, s := range services {
		svc, ok := s.(map[string]interface{})
		if !ok {
			continue
		}

		svcDomain, _ := svc["domain"].(string)
		if svcDomain != domain {
			continue
		}

		svcServices, ok := svc["services"].(map[string]interface{})
		if !ok {
			continue
		}

		serviceData, ok := svcServices[service].(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := serviceData["name"].(string)
		if name == "" {
			name = service
		}

		result := map[string]interface{}{
			"action":      actionName,
			"name":        name,
			"description": serviceData["description"],
			"fields":      serviceData["fields"],
			"target":      serviceData["target"],
		}

		client.PrintOutput(result, textMode, "")
		return nil
	}

	return fmt.Errorf("action '%s' not found", actionName)
}

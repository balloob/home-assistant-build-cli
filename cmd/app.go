package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:     "app",
	Aliases: []string{"apps", "addon"},
	Short:   "Manage Home Assistant apps",
	Long: `Manage Supervisor apps (formerly add-ons).

Requires a Home Assistant installation with Supervisor and an administrator
account. Use 'hab app list --available' to discover apps to install.`,
	GroupID: "other",
}

var appSlugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(newAppListCommand(), newAppGetCommand(), newAppConfigureCommand())
	for _, operation := range []string{"install", "update", "start", "stop", "restart", "uninstall"} {
		appCmd.AddCommand(newAppActionCommand(operation))
	}
}

func appSlugArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}
	if !appSlugPattern.MatchString(args[0]) {
		return fmt.Errorf("invalid app slug; use a slug from 'hab app list'")
	}
	return nil
}

func annotateAppCommand(cmd *cobra.Command, sideEffect string) {
	mergeSchemaAnnotation(cmd, SchemaAnnotation{
		SideEffect:   sideEffect,
		OutputMode:   "json_envelope",
		Capabilities: []string{"auth", "ws", "supervisor"},
		ResourceType: "app",
		GuideTopic:   "apps",
		InputSources: []string{"args", "flags"},
	})
}

func appRequest(cmd *cobra.Command, method, endpoint string, data map[string]any, timeout time.Duration) (any, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("--timeout must be positive")
	}
	signalCtx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, timeout)
	defer cancel()
	ws, err := getWSClientContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := ws.Close(); err != nil {
			noteWarning("App connection close failed.")
		}
	}()
	ws.Timeout = timeout
	// Supervisor operations can exceed its default proxy timeout.
	params := map[string]any{"endpoint": endpoint, "method": method, "timeout": nil}
	if data != nil {
		params["data"] = data
	}
	result, err := ws.SendCommandContext(ctx, "supervisor/api", params)
	if err != nil {
		if method != "get" {
			code, message, _ := classifyError(err)
			if code == client.ErrCodeTimeout || code == client.ErrCodeCancelled || code == client.ErrCodeConnectionError {
				return nil, &client.APIError{
					Code:              code,
					Message:           "App operation outcome is unknown. Supervisor may still complete it. Inspect the app state before retrying.",
					Category:          "unknown_outcome",
					Details:           map[string]any{"retryable": false, "cause": message},
					SuggestedFix:      "Inspect 'hab app list' and 'hab app get <slug>' before retrying.",
					SuggestedCommands: []string{"hab app list --json"},
					Transport:         "websocket",
				}
			}
		}
		return nil, err
	}
	if response, ok := result.(map[string]any); ok && response["result"] == "error" {
		return nil, &client.APIError{Code: client.ErrCodeAPIError, Message: fmt.Sprint(response["message"]), Transport: "websocket"}
	}
	return result, nil
}

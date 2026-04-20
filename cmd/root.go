// Package cmd defines the CLI commands for the hab tool using the Cobra
// framework. Each feature area (auth, entity, automation, etc.) has a parent
// command file and per-operation subcommand files.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	neturl "net/url"
	"os"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/home-assistant/hab/config"
	"github.com/home-assistant/hab/output"
	"github.com/home-assistant/hab/update"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgDir          string
	textMode        bool
	jsonMode        bool
	verbose         bool
	skipUpdateCheck bool
)

// ExitWithError signals that the program should exit with a non-zero code
var ExitWithError = false

var rootCmd = &cobra.Command{
	Use:   executableName(os.Args[0]),
	Short: "Home Assistant Builder - Build Home Assistant configurations",
	Long: `Home Assistant Builder (hab) is a CLI utility designed for LLMs
to build and manage Home Assistant configurations.

Interactive sessions default to human-readable text. Non-interactive sessions default to JSON.

Start with 'hab guide' for workflow-level guidance optimized for LLM and agent usage.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		resetExecutionMetadata()
		text, json, mode, err := determineOutputMode(cmd)
		if err != nil {
			return err
		}
		viper.Set("text", text)
		viper.Set("json", json)
		noteOutputMode(mode)

		// Set log level based on verbose flag
		if viper.GetBool("verbose") {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.WarnLevel)
		}

		log.WithFields(log.Fields{
			"url":     viper.GetString("url"),
			"text":    viper.GetBool("text"),
			"json":    viper.GetBool("json"),
			"verbose": viper.GetBool("verbose"),
			"config":  viper.GetString("config"),
		}).Debug("Configuration")

		// Check for updates (skip for update and version commands)
		checkUpdateOnStartup(cmd)
		return nil
	},
}

func executableName(arg0 string) string {
	if arg0 == "" {
		return "hab"
	}

	lastSeparator := strings.LastIndexAny(arg0, `/\`)
	if lastSeparator >= 0 && lastSeparator+1 < len(arg0) {
		return arg0[lastSeparator+1:]
	}

	return arg0
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		code, msg, details := classifyError(err)
		if viper.GetBool("json") {
			fmt.Println(output.FormatError(code, msg, details))
		} else {
			fmt.Fprintln(os.Stderr, msg)
		}
		ExitWithError = true
	}
}

// classifyError extracts a structured error code and user-facing message from err.
// It checks for known error types (APIError, auth sentinel) and falls back to UNKNOWN_ERROR.
func classifyError(err error) (code string, msg string, details map[string]any) {
	// Check for auth sentinel first (it's a plain error, not an APIError)
	if errors.Is(err, auth.ErrNotAuthenticated) {
		return client.ErrCodeAuthRequired, "Not authenticated. Run 'hab auth login' to authenticate.", map[string]any{
			"category":           "authentication",
			"retryable":          false,
			"likely_cause":       "No stored credentials were found for this Home Assistant instance.",
			"suggested_fix":      "Authenticate before running commands that require API access.",
			"suggested_commands": []string{"hab auth status --json", "hab auth login"},
		}
	}

	// Check for structured API errors (from REST or WebSocket)
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		details := apiErr.DetailsMap()
		if details == nil {
			details = map[string]any{}
		}
		if _, exists := details["category"]; !exists {
			details["category"] = "api"
		}
		if _, exists := details["suggested_fix"]; !exists {
			details["suggested_fix"] = defaultSuggestedFixForCode(apiErr.Code)
		}
		if _, exists := details["suggested_commands"]; !exists {
			if commands := defaultSuggestedCommandsForCode(apiErr.Code); len(commands) > 0 {
				details["suggested_commands"] = commands
			}
		}
		if len(details) == 0 {
			details = nil
		}
		return apiErr.Code, apiErr.Message, details
	}

	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
		return client.ErrCodeTimeout, "Operation timed out.", map[string]any{
			"category":           "timeout",
			"retryable":          true,
			"likely_cause":       "The Home Assistant API did not respond within the configured timeout.",
			"suggested_fix":      "Retry the command after Home Assistant is responsive.",
			"suggested_commands": []string{"hab system health --json"},
		}
	}

	if errors.Is(err, context.Canceled) {
		return client.ErrCodeCancelled, "Operation cancelled.", map[string]any{
			"category":      "cancellation",
			"retryable":     true,
			"suggested_fix": "Retry the command and confirm the prompt or provide required flags.",
		}
	}

	var netOpErr *net.OpError
	if errors.As(err, &netOpErr) {
		return client.ErrCodeConnectionError, "Connection failed.", map[string]any{
			"category":      "connection",
			"retryable":     true,
			"suggested_fix": "Verify Home Assistant URL/network connectivity and retry.",
		}
	}

	var urlErr *neturl.Error
	if errors.As(err, &urlErr) {
		return client.ErrCodeConnectionError, "Connection failed.", map[string]any{
			"category":      "connection",
			"retryable":     true,
			"suggested_fix": "Verify Home Assistant URL/network connectivity and retry.",
		}
	}

	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "invalid") || strings.Contains(lower, "parse") || strings.Contains(lower, "yaml") || strings.Contains(lower, "json") {
		return client.ErrCodeInputError, err.Error(), map[string]any{
			"category":      "input",
			"retryable":     false,
			"suggested_fix": "Check command input/flags and provide valid JSON/YAML payloads.",
		}
	}

	// Fallback
	return client.ErrCodeUnknownError, err.Error(), map[string]any{
		"category":      "unknown",
		"retryable":     false,
		"suggested_fix": "Inspect the error details and retry with --verbose for additional context.",
	}
}

func defaultSuggestedFixForCode(code string) string {
	switch code {
	case client.ErrCodeAuthenticationError, client.ErrCodeAuthRequired:
		return "Authenticate or refresh credentials, then retry."
	case client.ErrCodePermissionDenied:
		return "Use credentials with sufficient privileges for this command."
	case client.ErrCodeNotFound:
		return "List resources first and retry with a valid identifier from current output."
	case client.ErrCodeValidationError, client.ErrCodeInputError:
		return "Correct the command input and retry."
	case client.ErrCodeConnectionError:
		return "Verify URL/network availability and retry."
	case client.ErrCodeTimeout:
		return "Retry after Home Assistant is responsive."
	case client.ErrCodeConfirmationRequired:
		return "Use --force only after validating the target, or rerun interactively to confirm the prompt."
	case client.ErrCodeCancelled:
		return "Retry the command and confirm the prompt or use --force when appropriate."
	default:
		return "Retry with --verbose and inspect command output for additional context."
	}
}

func defaultSuggestedCommandsForCode(code string) []string {
	switch code {
	case client.ErrCodeAuthenticationError, client.ErrCodeAuthRequired:
		return []string{"hab auth status --json", "hab auth login"}
	case client.ErrCodePermissionDenied:
		return []string{"hab auth status --json"}
	case client.ErrCodeConnectionError, client.ErrCodeTimeout:
		return []string{"hab system health --json"}
	case client.ErrCodeConfirmationRequired:
		return []string{"hab schema --json"}
	default:
		return nil
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.AddTemplateFunc("formatExamples", formatExamples)

	// Silence usage and errors - we handle error display ourselves
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.SetUsageTemplate(usageTemplate)

	// Disable shell completion command (not useful for LLM usage)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add command groups for better organization
	rootCmd.AddGroup(&cobra.Group{ID: "start", Title: "Getting Started:"})
	rootCmd.AddGroup(&cobra.Group{ID: "registry", Title: "Registry:"})
	rootCmd.AddGroup(&cobra.Group{ID: "automation", Title: "Automation:"})
	rootCmd.AddGroup(&cobra.Group{ID: "dashboard", Title: "Dashboard:"})
	rootCmd.AddGroup(&cobra.Group{ID: "other", Title: "Other:"})

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgDir, "config", "", "Path to config directory (default: ~/.config/home-assistant-builder)")
	rootCmd.PersistentFlags().BoolVar(&jsonMode, "json", false, "Use JSON output instead of human-readable text")
	rootCmd.PersistentFlags().BoolVar(&textMode, "text", false, "Use human-readable text output")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Show verbose output")
	rootCmd.PersistentFlags().BoolVar(&skipUpdateCheck, "skip-update-check", false, "Skip automatic update check on startup")

	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	viper.BindPFlag("text", rootCmd.PersistentFlags().Lookup("text"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("skip-update-check", rootCmd.PersistentFlags().Lookup("skip-update-check"))

	// Shell completions
	rootCmd.RegisterFlagCompletionFunc("json", boolCompletions)
	rootCmd.RegisterFlagCompletionFunc("text", boolCompletions)
	rootCmd.RegisterFlagCompletionFunc("verbose", boolCompletions)
	rootCmd.RegisterFlagCompletionFunc("skip-update-check", boolCompletions)
	rootCmd.MarkPersistentFlagDirname("config")
}

func initConfig() {
	// Set environment variable prefix
	viper.SetEnvPrefix("HAB")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Bind specific environment variables
	viper.BindEnv("url", "HAB_URL")
	viper.BindEnv("token", "HAB_TOKEN")
	viper.BindEnv("refresh-token", "HAB_REFRESH_TOKEN")
	viper.BindEnv("skip-update-check", "HAB_SKIP_UPDATE_CHECK")

	// Set defaults
	config.InitDefaults()

	// Read config file if it exists
	if cfgDir != "" {
		viper.SetConfigFile(config.GetConfigPath(cfgDir))
	} else {
		configDir := config.GetConfigDir("")
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("json")
	}

	// Read config file (ignore errors if not found)
	if err := viper.ReadInConfig(); err == nil {
		log.WithField("configfile", viper.ConfigFileUsed()).Debug("Using config file")
	}
}

func boolCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
}

// checkUpdateOnStartup checks for updates once per day and prints a notice if available
func checkUpdateOnStartup(cmd *cobra.Command) {
	if !isInteractiveOutput() {
		return
	}
	// Skip for certain commands
	cmdName := cmd.Name()
	if cmdName == "update" || cmdName == "version" || cmdName == "help" || cmdName == "guide" || cmdName == "schema" {
		return
	}

	// Skip if flag is set or env var is set
	if viper.GetBool("skip-update-check") {
		return
	}

	// Skip if version is dev (development build)
	if Version == "" || Version == "dev" {
		return
	}

	configDir := viper.GetString("config")

	// Check if we need to check for updates (once per day)
	if !update.NeedsCheck(configDir) {
		// Load cached check to see if we should show notice
		check, err := update.LoadUpdateCheck(configDir)
		if err == nil && check != nil {
			check.CurrentVersion = Version
			if update.HasUpdate(check) {
				update.PrintUpdateNotice(check)
			}
		}
		return
	}

	// Perform update check in background (don't block startup)
	go func() {
		check, err := update.CheckForUpdate(configDir, Version)
		if err != nil {
			log.WithError(err).Debug("Failed to check for updates")
			return
		}

		if update.HasUpdate(check) {
			update.PrintUpdateNotice(check)
		}
	}()
}

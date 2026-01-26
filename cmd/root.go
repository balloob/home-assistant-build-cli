package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/config"
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
	Use:   path.Base(os.Args[0]),
	Short: "Home Assistant Builder - Build Home Assistant configurations",
	Long: `Home Assistant Builder (hab) is a CLI utility designed for LLMs
to build and manage Home Assistant configurations.

Output is human-readable text by default. Use --json for machine-parseable JSON output.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Handle --json flag: if set, override text mode to false
		if viper.GetBool("json") {
			viper.Set("text", false)
		}

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
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Don't print auth errors again - warning was already shown
		if !errors.Is(err, auth.ErrNotAuthenticated) {
			fmt.Fprintln(os.Stderr, err)
		}
		ExitWithError = true
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Silence usage and errors - we handle error display ourselves
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgDir, "config", "", "Path to config directory (default: ~/.config/home-assistant-builder)")
	rootCmd.PersistentFlags().BoolVar(&jsonMode, "json", false, "Use JSON output instead of human-readable text")
	rootCmd.PersistentFlags().BoolVar(&textMode, "text", true, "Use human-readable text output (default)")
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
	// Skip for certain commands
	cmdName := cmd.Name()
	if cmdName == "update" || cmdName == "version" || cmdName == "help" {
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

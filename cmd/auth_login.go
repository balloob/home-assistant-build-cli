package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	loginToken       bool
	loginURL         string
	loginAccessToken string
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Home Assistant",
	Long:  `Authenticate with Home Assistant using OAuth flow or a long-lived access token.`,
	RunE:  runAuthLogin,
}

func init() {
	authCmd.AddCommand(authLoginCmd)

	authLoginCmd.Flags().BoolVar(&loginToken, "token", false, "Use long-lived access token instead of OAuth")
	authLoginCmd.Flags().StringVar(&loginURL, "url", "", "Home Assistant URL")
	authLoginCmd.Flags().StringVar(&loginAccessToken, "access-token", "", "Long-lived access token (non-interactive mode)")
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	configDir := viper.GetString("config")
	textMode := viper.GetBool("text")
	manager := auth.NewManager(configDir)

	if loginToken {
		return loginWithToken(manager, textMode)
	}
	return loginWithOAuth(manager, textMode)
}

// getURLWithDiscovery attempts to discover Home Assistant servers on the network
// and presents a selection menu. Returns the selected or manually entered URL.
func getURLWithDiscovery() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Searching for Home Assistant servers...")
	servers, err := auth.DiscoverServers(3 * time.Second)

	if err != nil || len(servers) == 0 {
		// No servers found or discovery failed, fall back to manual entry
		if err != nil {
			fmt.Printf("Discovery failed: %v\n", err)
		} else {
			fmt.Println("No Home Assistant servers found on the network.")
		}
		fmt.Print("\nHome Assistant URL: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read URL: %w", err)
		}
		return strings.TrimSpace(input), nil
	}

	// Display discovered servers
	fmt.Printf("\nFound %d Home Assistant server(s):\n\n", len(servers))
	for i, server := range servers {
		fmt.Printf("  [%d] %s\n", i+1, auth.FormatServerDisplay(server))
	}
	fmt.Printf("  [%d] Enter URL manually\n", len(servers)+1)

	// Get user selection
	fmt.Printf("\nSelect server [1-%d]: ", len(servers)+1)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	selection := strings.TrimSpace(input)
	idx, err := strconv.Atoi(selection)
	if err != nil || idx < 1 || idx > len(servers)+1 {
		return "", fmt.Errorf("invalid selection: %s", selection)
	}

	// Manual entry option
	if idx == len(servers)+1 {
		fmt.Print("Home Assistant URL: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read URL: %w", err)
		}
		return strings.TrimSpace(input), nil
	}

	return servers[idx-1].URL, nil
}

func loginWithToken(manager *auth.Manager, textMode bool) error {
	// Get URL
	url := loginURL
	if url == "" {
		var err error
		url, err = getURLWithDiscovery()
		if err != nil {
			return err
		}
	}

	// Get access token
	accessToken := loginAccessToken
	if accessToken == "" {
		fmt.Print("Long-lived access token: ")
		tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
		accessToken = string(tokenBytes)
	}

	// Validate the token by making a test request
	restClient := client.NewRestClient(url, accessToken)
	config, err := restClient.GetConfig()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save credentials
	creds := &auth.Credentials{
		URL:         url,
		AccessToken: accessToken,
	}
	if err := manager.Save(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	// Output result
	locationName := "Home Assistant"
	if name, ok := config["location_name"].(string); ok {
		locationName = name
	}

	result := map[string]interface{}{
		"url":           url,
		"location_name": locationName,
		"version":       config["version"],
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Successfully authenticated to %s", locationName))
	return nil
}

func loginWithOAuth(manager *auth.Manager, textMode bool) error {
	// Get URL
	url := loginURL
	if url == "" {
		var err error
		url, err = getURLWithDiscovery()
		if err != nil {
			return err
		}
	}

	// Run OAuth flow
	creds, err := auth.RunOAuthFlow(url)
	if err != nil {
		return fmt.Errorf("OAuth flow failed: %w", err)
	}

	// Validate and get info
	restClient := client.NewRestClient(creds.URL, creds.AccessToken)
	config, err := restClient.GetConfig()
	if err != nil {
		return fmt.Errorf("authentication validation failed: %w", err)
	}

	// Save credentials
	if err := manager.Save(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	// Output result
	locationName := "Home Assistant"
	if name, ok := config["location_name"].(string); ok {
		locationName = name
	}

	result := map[string]interface{}{
		"url":           url,
		"location_name": locationName,
		"version":       config["version"],
	}

	client.PrintSuccess(result, textMode, fmt.Sprintf("Successfully authenticated to %s", locationName))
	return nil
}

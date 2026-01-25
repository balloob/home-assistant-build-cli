package cmd

import (
	"time"

	"github.com/home-assistant/hab/auth"
	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var discoverTimeout int

var authDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover Home Assistant servers on the network",
	Long:  `Search for Home Assistant servers on the local network using mDNS/DNS-SD.`,
	RunE:  runAuthDiscover,
}

func init() {
	authCmd.AddCommand(authDiscoverCmd)

	authDiscoverCmd.Flags().IntVar(&discoverTimeout, "timeout", 3, "Discovery timeout in seconds")
}

func runAuthDiscover(cmd *cobra.Command, args []string) error {
	textMode := viper.GetBool("text")

	servers, err := auth.DiscoverServers(time.Duration(discoverTimeout) * time.Second)
	if err != nil {
		return err
	}

	// Convert to output format
	results := make([]map[string]interface{}, len(servers))
	for i, server := range servers {
		results[i] = map[string]interface{}{
			"name":    server.Name,
			"url":     server.URL,
			"version": server.Version,
			"uuid":    server.UUID,
		}
	}

	if textMode {
		if len(results) == 0 {
			client.PrintOutput(nil, textMode, "No Home Assistant servers found on the network.")
		} else {
			for _, server := range servers {
				client.PrintOutput(nil, textMode, auth.FormatServerDisplay(server))
			}
		}
	} else {
		client.PrintOutput(results, textMode, "")
	}

	return nil
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var esphomeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ESPHome devices",
	Long:  `List all configured ESPHome devices with their status, platform, and version info.`,
	RunE:  runESPHomeList,
}

func init() {
	esphomeCmd.AddCommand(esphomeListCmd)
}

func runESPHomeList(cmd *cobra.Command, args []string) error {
	textMode := viper.GetBool("text")

	esClient, err := getESPHomeClient()
	if err != nil {
		return err
	}

	devices, err := esClient.GetDevices()
	if err != nil {
		return err
	}

	ping, _ := esClient.GetPing() // best-effort, don't fail if ping fails

	if textMode {
		if len(devices.Configured) == 0 && len(devices.Importable) == 0 {
			fmt.Println("No ESPHome devices found.")
			return nil
		}

		if len(devices.Configured) > 0 {
			fmt.Printf("Configured devices (%d):\n\n", len(devices.Configured))
			for _, d := range devices.Configured {
				status := "unknown"
				if ping != nil {
					if s, ok := ping[d.Configuration]; ok && s != nil {
						if *s {
							status = "online"
						} else {
							status = "offline"
						}
					}
				}

				name := d.FriendlyName
				if name == "" {
					name = d.Name
				}
				fmt.Printf("  %s (%s)\n", name, d.Configuration)
				fmt.Printf("    Platform: %s  Status: %s\n", d.TargetPlatform, status)
				if d.DeployedVersion != "" || d.CurrentVersion != "" {
					parts := []string{}
					if d.DeployedVersion != "" {
						parts = append(parts, "deployed: "+d.DeployedVersion)
					}
					if d.CurrentVersion != "" {
						parts = append(parts, "current: "+d.CurrentVersion)
					}
					fmt.Printf("    Version: %s\n", strings.Join(parts, ", "))
				}
				if d.Address != "" {
					fmt.Printf("    Address: %s\n", d.Address)
				}
				fmt.Println()
			}
		}

		if len(devices.Importable) > 0 {
			fmt.Printf("Importable devices (%d):\n\n", len(devices.Importable))
			for _, d := range devices.Importable {
				name := d.FriendlyName
				if name == "" {
					name = d.Name
				}
				ignored := ""
				if d.Ignored {
					ignored = " (ignored)"
				}
				fmt.Printf("  %s%s\n", name, ignored)
				if d.ProjectName != "" {
					fmt.Printf("    Project: %s %s\n", d.ProjectName, d.ProjectVersion)
				}
				fmt.Println()
			}
		}

		return nil
	}

	// JSON mode: return combined data with ping status merged in
	type deviceWithStatus struct {
		client.ESPHomeDevice
		Online *bool `json:"online"`
	}

	var enriched []deviceWithStatus
	for _, d := range devices.Configured {
		dws := deviceWithStatus{ESPHomeDevice: d}
		if ping != nil {
			if s, ok := ping[d.Configuration]; ok {
				dws.Online = s
			}
		}
		enriched = append(enriched, dws)
	}

	result := map[string]interface{}{
		"configured": enriched,
		"importable": devices.Importable,
	}

	client.PrintOutput(result, false, "")
	return nil
}

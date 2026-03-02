package cmd

import (
	"strings"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
)

var logsLines int

var systemLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get error logs",
	Long:  `Get the Home Assistant error log.`,
	RunE:  runSystemLogs,
}

func init() {
	systemCmd.AddCommand(systemLogsCmd)
	systemLogsCmd.Flags().IntVarP(&logsLines, "lines", "n", 100, "Number of lines to show")
}

func runSystemLogs(cmd *cobra.Command, args []string) error {
	textMode := getTextMode()

	restClient, err := getRESTClient()
	if err != nil {
		return err
	}

	logContent, err := restClient.GetErrorLog()
	if err != nil {
		return err
	}

	// Get last N lines
	lines := strings.Split(strings.TrimSpace(logContent), "\n")
	if len(lines) > logsLines {
		lines = lines[len(lines)-logsLines:]
	}

	client.PrintOutput(strings.Join(lines, "\n"), textMode, "")
	return nil
}

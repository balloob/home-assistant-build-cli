package cmd

import (
	"runtime"

	"github.com/home-assistant/hab/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is set by main.go from ldflags
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Show version information",
	Long:    `Display the current version of hab and build information.`,
	Run:     runVersion,
	GroupID: "other",
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	textMode := viper.GetBool("text")

	versionInfo := map[string]interface{}{
		"version":   Version,
		"go":        runtime.Version(),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"compiler":  runtime.Compiler,
	}

	if textMode {
		client.PrintOutput(nil, true, "hab version "+Version+" ("+runtime.GOOS+"/"+runtime.GOARCH+")")
	} else {
		client.PrintSuccess(versionInfo, false, "")
	}
}

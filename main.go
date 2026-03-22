package main

import (
	"os"

	"github.com/home-assistant/hab/cmd"
	"github.com/home-assistant/hab/update"
)

// version is set at build time via ldflags
var version = "dev"

func main() {
	if handled, exitCode := update.MaybeRunWindowsUpdateHelper(os.Args); handled {
		os.Exit(exitCode)
	}

	cmd.Version = version
	cmd.Execute()
	if cmd.ExitWithError {
		os.Exit(1)
	}
}

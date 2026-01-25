package main

import (
	"os"

	"github.com/home-assistant/hab/cmd"
)

// version is set at build time via ldflags
var version = "dev"

func main() {
	cmd.Version = version
	cmd.Execute()
	if cmd.ExitWithError {
		os.Exit(1)
	}
}

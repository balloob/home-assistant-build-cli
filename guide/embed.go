// Package guide provides embedded markdown guides accessible at runtime.
package guide

import "embed"

// Guides holds the embedded markdown files from the guide/ directory.
//
//go:embed *.md
var Guides embed.FS

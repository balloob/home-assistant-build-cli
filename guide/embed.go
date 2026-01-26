package guide

import "embed"

//go:embed *.md
var Guides embed.FS

// Get returns the content of a guide by name (without .md extension)
func Get(name string) (string, error) {
	content, err := Guides.ReadFile(name + ".md")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

package cmd

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// These are path-specific corrections from the 1.7.1 implementation audit.
// Apply before generic preview wiring so misclassified mutations cannot run
// their implementation when --plan is requested.
var auditedSideEffects = map[string]string{
	"automation create-from-blueprint": "write",
	"dashboard save-config":            "write",
	"esphome build":                    "write",
	"esphome upload":                   "write",
	"esphome config-patch":             "write",
	"esphome config-write":             "write",
	"esphome serial probe":             "write", // esptool can reset the target MCU.
	"esphome context clear":            "destructive",
	"esphome context use":              "write",
	"repairs ignore":                   "write",
	"repairs unignore":                 "write",
	"scene activate":                   "write",
	"thread set-preferred":             "write",
}

var auditedCapabilities = map[string][]string{
	"auth discover":                            {"mdns"},
	"auth login":                               {"local", "http", "mdns"},
	"auth refresh":                             {"local", "http"},
	"automation trace":                         {"auth", "ws"},
	"blueprint get":                            {"auth", "ws"},
	"blueprint import":                         {"auth", "ws"},
	"blueprint list":                           {"auth", "ws"},
	"capability probe":                         {"auth", "rest", "ws", "esphome"},
	"dashboard guide":                          {"local"},
	"entity get":                               {"auth", "rest", "ws"},
	"entity history":                           {"auth", "rest"},
	"entity logbook":                           {"auth", "rest"},
	"esphome catalog search":                   {"http"},
	"esphome catalog show":                     {"http"},
	"esphome context clear":                    {"local"},
	"esphome context show":                     {"local"},
	"esphome context use":                      {"local"},
	"esphome migrate tasmota-template analyze": {"local"},
	"esphome serial ports":                     {"local", "serial"},
	"esphome serial probe":                     {"local", "serial"},
	"esphome serial erase-flash":               {"local", "serial"},
	"helper types":                             {"local"},
	"help":                                     {"local"},
	"integration disable":                      {"auth", "ws"},
	"integration enable":                       {"auth", "ws"},
	"integration get":                          {"auth", "ws"},
	"integration list":                         {"auth", "ws"},
	"notification list":                        {"auth", "ws"},
	"scene list":                               {"auth", "ws"},
	"system health":                            {"auth", "ws"},
	"system updates":                           {"auth", "ws"},
	"update":                                   {"local", "http"},
}

func applyAuditedMetadata(cmd *cobra.Command) {
	path := strings.TrimPrefix(schemaPath(cmd), "hab ")
	ann := getSchemaAnnotation(cmd)
	if effect := auditedSideEffects[path]; effect != "" {
		ann.SideEffect = effect
	}
	if caps, ok := auditedCapabilities[path]; ok {
		ann.Capabilities = slices.Clone(caps)
	}
	if path == "esphome validate" {
		ann.OutputMode = "ndjson_stream"
	}
	if path == "help" {
		ann.OutputMode = "text"
	}
	switch path {
	case "device entities":
		ann.ResourceType = "entity"
	case "thread add", "thread get", "thread set-preferred":
		ann.ResourceType = "thread_dataset"
	case "schema":
		ann.ResourceType = "command"
		ann.Args = []SchemaPositionalArg{{Name: "command_path", Raw: "[command path]", Variadic: true}}
	}
	if slices.Contains([]string{"esphome build", "esphome upload", "esphome run", "esphome logs", "esphome validate", "esphome info", "esphome config-patch", "esphome update"}, path) {
		// These resolve the positional value, --device, or saved device context.
		ann.Args = []SchemaPositionalArg{{Name: "configuration", Raw: "[configuration]"}}
		if !slices.Contains(ann.InputSources, "saved_context") {
			ann.InputSources = append(ann.InputSources, "saved_context")
		}
	}
	setSchemaAnnotation(cmd, ann)
}

func commandTransports(cmd *cobra.Command) []string {
	caps := getSchemaAnnotation(cmd).Capabilities
	if override, ok := auditedCapabilities[strings.TrimPrefix(schemaPath(cmd), "hab ")]; ok {
		caps = override
	}
	var transports []string
	for _, cap := range caps {
		if slices.Contains([]string{"local", "rest", "ws", "http", "mdns", "serial", "esphome"}, cap) {
			transports = append(transports, cap)
		}
	}
	return transports
}

func commandPreview(cmd *cobra.Command) string {
	if cmd.Flags().Lookup("plan") == nil && cmd.Flags().Lookup("dry-run") == nil {
		return "none"
	}
	if schemaPath(cmd) == "hab dashboard patch" {
		return "live_diff"
	}
	if cmd.Annotations["hab/generic-plan"] == "true" {
		return "static"
	}
	// Legacy command-specific planners vary in whether they resolve IDs live;
	// none imply server validation or an actual diff.
	return "command_plan"
}

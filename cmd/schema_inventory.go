package cmd

import "strings"

// Audited JSON data shapes, independent of verb inference. "any" is an
// explicitly opaque API/flag-dependent result, not a fabricated object schema.
// Factory families are enumerated here so adding any executable requires review.
func auditedDataShapes() map[string]string {
	result := map[string]string{}
	groups := map[string]string{
		"null": `app configure
app install
app update
app start
app stop
app restart
app uninstall
auth logout
automation run
backup delete
calendar create
calendar delete
category delete
dashboard delete
dashboard save-config
device delete
esphome config-write
esphome context clear
event fire
marketplace accept-warning
marketplace clear-new
marketplace custom add
marketplace custom remove
marketplace ignore
marketplace install
marketplace refresh
marketplace set-beta
marketplace set-version
marketplace uninstall
notification create
notification dismiss
person delete
repairs ignore
repairs unignore
scene activate
script run
system restart
thread delete
thread set-preferred
todo add
todo complete
todo remove
todo uncomplete
todo update
zone delete`,
		"array": `app list
auth context list
action data
action list
automation list
category list
dashboard list
device entities
device list
id list
diagnostics list
entity history
entity list
entity logbook
entity search
esphome boards
esphome catalog search
esphome list
esphome serial ports
event list
guide list
helper list
helper types
integration list
marketplace list
notification list
person list
repairs list
scene list
script list
system updates
todo items
todo lists
zone list`,
		"object": `app get
action docs
auth context use
auth login
auth refresh
auth status
blueprint get
blueprint import
capability probe
category assign
category create
category remove
category update
dashboard create
dashboard get
dashboard guide
dashboard patch
dashboard update
device get
device update
entity disable
entity enable
entity get
entity rename
entity update
esphome catalog show
esphome config-patch
esphome config-read
esphome context show
esphome context use
esphome import
esphome info
esphome migrate tasmota-template analyze
esphome migrate tasmota-template create
esphome serial erase-flash
esphome serial probe
esphome update
helper delete
integration disable
integration enable
integration get
integration reload
marketplace github-connect
overview
person create
person get
person update
schema
search related
system info
template render
update
version
zone create
zone update`,
		"string": `system logs`,
		"any": `action call
auth discover
automation create-from-blueprint
automation trace
backup agents
backup config get
backup config update
backup create
backup get
backup list
backup restore
blueprint delete
blueprint list
calendar list
diagnostics get
energy info
energy prefs get
energy prefs set
energy solar-forecast
energy validate
esphome create
guide
help
label assign
label remove
marketplace critical acknowledge
marketplace critical list
marketplace custom detect
marketplace get
marketplace info
marketplace releases
marketplace removed
network configure
network get
network url
system config-check
system health
thread add
thread get
thread list`,
		"stream": `esphome build
esphome logs
esphome run
esphome upload
esphome validate`,
	}
	management := map[string]string{
		"hab app logs":                        "string",
		"hab app rebuild":                     "any",
		"hab app refresh":                     "any",
		"hab app repository add":              "any",
		"hab app repository list":             "array",
		"hab app repository remove":           "any",
		"hab app security":                    "any",
		"hab app stats":                       "object",
		"hab app validate":                    "object",
		"hab application-credential config":   "object",
		"hab application-credential create":   "object",
		"hab application-credential delete":   "null",
		"hab application-credential entry":    "object",
		"hab application-credential list":     "array",
		"hab assist devices":                  "array",
		"hab assist exposure defaults get":    "object",
		"hab assist exposure defaults set":    "null",
		"hab assist exposure list":            "object",
		"hab assist exposure set":             "object",
		"hab assist languages":                "object",
		"hab assist pipeline create":          "object",
		"hab assist pipeline delete":          "null",
		"hab assist pipeline get":             "object",
		"hab assist pipeline list":            "object",
		"hab assist pipeline set-preferred":   "null",
		"hab assist pipeline update":          "object",
		"hab assist runs get":                 "object",
		"hab assist runs list":                "object",
		"hab backup download":                 "object",
		"hab backup upload":                   "object",
		"hab diagnostics download":            "object",
		"hab host datadisk list":              "any",
		"hab host datadisk move":              "any",
		"hab host disks list":                 "any",
		"hab host disks usage":                "any",
		"hab host info":                       "object",
		"hab host options":                    "any",
		"hab host os info":                    "object",
		"hab host os sync":                    "any",
		"hab host os update":                  "any",
		"hab host reboot":                     "any",
		"hab host shutdown":                   "any",
		"hab integration delete":              "object",
		"hab integration flow abort":          "object",
		"hab integration flow get":            "object",
		"hab integration flow list":           "array",
		"hab integration flow start":          "object",
		"hab integration flow step":           "object",
		"hab integration options abort":       "object",
		"hab integration options get":         "object",
		"hab integration options start":       "object",
		"hab integration options step":        "object",
		"hab integration subentry delete":     "null",
		"hab integration subentry flow abort": "object",
		"hab integration subentry flow get":   "object",
		"hab integration subentry flow start": "object",
		"hab integration subentry flow step":  "object",
		"hab integration subentry list":       "array",
		"hab integration subentry update":     "null",
		"hab integration types":               "array",
		"hab integration update":              "object",
		"hab media browse":                    "object",
		"hab media delete":                    "null",
		"hab media resolve":                   "object",
		"hab media search":                    "object",
		"hab media upload":                    "object",
		"hab mount create":                    "any",
		"hab mount delete":                    "any",
		"hab mount list":                      "object",
		"hab mount options":                   "any",
		"hab mount reload":                    "any",
		"hab mount update":                    "any",
		"hab network interfaces list":         "object",
		"hab network interfaces scan":         "any",
		"hab network interfaces update":       "any",
		"hab repairs flow abort":              "object",
		"hab repairs flow get":                "object",
		"hab repairs flow start":              "object",
		"hab repairs flow step":               "object",
		"hab repairs get":                     "object",
		"hab statistics adjust-sum":           "null",
		"hab statistics clear":                "null",
		"hab statistics metadata list":        "array",
		"hab statistics metadata update":      "null",
		"hab statistics query":                "object",
		"hab statistics refresh-issues":       "null",
		"hab statistics summary":              "object",
		"hab statistics validate":             "object",
		"hab system config detect":            "object",
		"hab system config get":               "object",
		"hab system config update":            "null",
		"hab user create":                     "object",
		"hab user credentials create":         "null",
		"hab user delete":                     "null",
		"hab user list":                       "array",
		"hab user password":                   "null",
		"hab user update":                     "object",
		"hab user username":                   "null",
	}
	for path, shape := range management {
		result[path] = shape
	}

	for shape, paths := range groups {
		for _, path := range strings.Split(paths, "\n") {
			result["hab "+path] = shape
		}
	}
	for _, family := range []string{"area", "floor", "label", "automation", "scene", "script", "dashboard view", "dashboard section", "dashboard badge", "dashboard card", "automation action", "automation condition", "automation trigger", "script action"} {
		for _, op := range []string{"get", "create", "update"} {
			result["hab "+family+" "+op] = "object"
		}
		result["hab "+family+" delete"] = "null"
		result["hab "+family+" list"] = "array"
	}
	// Config writes return HA's acknowledgement, not the supplied config.
	for _, family := range []string{"automation", "scene", "script"} {
		for _, op := range []string{"create", "update"} {
			result["hab "+family+" "+op] = "any"
		}
	}
	for _, helper := range strings.Fields("counter derivative group input-boolean input-button input-datetime input-number input-select input-text integration local-calendar local-todo min-max schedule statistics template threshold timer utility-meter") {
		result["hab helper "+helper+" list"] = "array"
		result["hab helper "+helper+" create"] = "object"
		result["hab helper "+helper+" delete"] = "object"
	}
	return result
}

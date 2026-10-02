package cmd

import "strings"

// Audited JSON data shapes, independent of verb inference. "any" is an
// explicitly opaque API/flag-dependent result, not a fabricated object schema.
// Factory families are enumerated here so adding any executable requires review.
func auditedDataShapes() map[string]string {
	result := map[string]string{}
	groups := map[string]string{
		"null": `auth logout
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
		"array": `action data
action list
automation list
category list
dashboard list
device entities
device list
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
		"object": `action docs
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
entity disable
entity enable
entity get
entity rename
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

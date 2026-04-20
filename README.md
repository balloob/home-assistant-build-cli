# Home Assistant Builder (hab)

A CLI utility designed for LLMs to build and manage Home Assistant configurations.

_Vibe coded, use at own risk._

## Installation

### From Source

```bash
go install github.com/balloob/home-assistant-build-cli@latest
```

Or build locally:

```bash
git clone https://github.com/balloob/home-assistant-build-cli
cd home-assistant-build-cli
go build -o hab .
```

## Quick Start

If you are driving `hab` from an LLM or automation, start with:

```bash
hab guide
hab guide list --json
hab schema overview --json
hab capability probe --json
```

For write workflows, inspect the command contract and preview the mutation first:

```bash
hab schema area create --json
hab area create "Kitchen" --plan --json
```

## Authentication

`hab` supports OAuth login, long-lived access tokens, stored encrypted credentials, and Home Assistant Supervisor credentials.

### Common auth flows

```bash
# Discover Home Assistant instances on the local network
hab auth discover
hab auth discover --timeout 5

# Authenticate using OAuth
hab auth login

# Authenticate using a long-lived access token
hab auth login --token --url http://homeassistant.local:8123 --access-token "your_token"

# Check authentication status
hab auth status

# Refresh OAuth credentials if needed
hab auth refresh

# Remove stored credentials
hab auth logout
```

### Auth behavior

- If `--url` is omitted during `hab auth login`, `hab` can discover servers and prompt for selection.
- Stored credentials are encrypted and saved in `credentials.json`.
- Credential resolution order is:
1. `HAB_URL` + `HAB_TOKEN` or `HAB_URL` + `HAB_REFRESH_TOKEN`
2. Encrypted stored credentials
3. `SUPERVISOR_TOKEN` fallback when running inside a Home Assistant add-on or app context
- OAuth credentials auto-refresh when needed.

## Common Usage Patterns

### Output modes

Interactive sessions default to human-readable text. Non-interactive sessions default to JSON.

```bash
hab entity get light.living_room --json
hab entity get light.living_room --text
```

### Common list flags

Many `list` commands support:

- `--count` / `-c`: return only the count
- `--brief` / `-b`: return only the ID and name fields
- `--limit` / `-n`: limit the number of returned items

Examples:

```bash
hab entity list --count
hab device list --brief --limit 5
hab automation list --limit 10
hab script list --brief
hab area list --limit 2
```

### Safe mutation previews

Many mutating commands support `--plan` and `--dry-run`.

```bash
hab area create "Kitchen" --plan --json
hab dashboard card create my-dashboard --entity light.kitchen --plan --json
hab system restart --plan --json
hab thread delete <dataset_id> --plan --json
```

### Confirmation and force

Destructive commands often require confirmation. Use `--force` to skip the prompt.

```bash
hab area delete <area_id> --force
hab device delete <device_id> --force
hab backup restore <backup_id> --agent backup.local --force
```

In non-interactive mode, commands that need confirmation return structured confirmation-required or cancelled errors instead of prompting.

## Discovery And Registry

### Overview

`hab overview` returns a high-level snapshot of the instance, including counts for floors, areas, devices, entities, automations, scripts, dashboards, labels, and helpers.

```bash
hab overview
hab overview --json
```

### Entities

```bash
# List entities
hab entity list
hab entity list --domain light
hab entity list --area kitchen
hab entity list --floor ground_floor
hab entity list --device abc123
hab entity list --device-class temperature
hab entity list --brief --limit 10

# Get entity state and registry data
hab entity get light.living_room
hab entity get light.living_room --device
hab entity get light.living_room --related

# Search entities by text
hab entity search motion

# Inspect state history
hab entity history sensor.temperature
hab entity history sensor.temperature --start "2025-01-01T00:00:00Z" --end "2025-01-02T00:00:00Z"

# View logbook entries
hab entity logbook light.living_room
hab entity logbook light.living_room --start "2024-01-01T00:00:00Z" --end "2024-01-02T00:00:00Z"

# Registry mutations
hab entity rename sensor.outdoor_temp "Outdoor Temperature"
hab entity disable sensor.outdoor_temp
hab entity enable sensor.outdoor_temp
```

### Devices

```bash
# List and filter devices
hab device list
hab device list --area kitchen
hab device list --floor ground_floor
hab device list --brief --limit 5

# Inspect a device and its entities
hab device get <device_id>
hab device get <device_id> --related
hab device entities <device_id>

# Delete a device
hab device delete <device_id> --plan --json
hab device delete <device_id> --force
```

### Areas, Floors, Labels, Zones, Persons

```bash
# Areas
hab area list
hab area list --floor ground_floor
hab area get <area_id>
hab area get <area_id> --related
hab area create "Kitchen"
hab area update <area_id> --name "Main Kitchen"
hab area delete <area_id> --force

# Floors
hab floor list
hab floor get <floor_id>
hab floor get <floor_id> --related
hab floor create "Ground Floor" --level 0
hab floor update <floor_id> --name "Ground Level"
hab floor delete <floor_id> --force

# Labels
hab label list
hab label get <label_id>
hab label get <label_id> --related
hab label create "Battery" --color red
hab label update <label_id> --name "Low Battery"
hab label assign <label_id> sensor.battery_level
hab label remove <label_id> sensor.battery_level
hab label delete <label_id> --force

# Zones
hab zone list
hab zone create "Office" --latitude 37.7749 --longitude -122.4194 --radius 100
hab zone update <zone_id> --name "HQ"
hab zone delete <zone_id> --force

# Persons
hab person list
hab person get <person_id>
hab person create "John Doe"
hab person update <person_id> --name "Jane Doe"
hab person delete <person_id> --force
```

### Search relationships

```bash
hab search related entity light.living_room
hab search related device <device_id>
hab search related area <area_id>
```

## Actions, Automations, Scripts, Scenes, Templates, Categories, Blueprints

### Actions

`action` maps to Home Assistant services.

```bash
# Discover available actions and docs
hab action list
hab action list light
hab action docs homeassistant.turn_on
hab action data

# Call actions
hab action call light.turn_on --entity light.living_room
hab action call climate.set_temperature --entity climate.living_room --data '{"temperature": 22}'
hab action call weather.get_forecasts --entity weather.home --data '{"type":"daily"}' --return-response
hab action call light.turn_off --area living_room
```

Flags accepted by `hab action call` include `--action`, `--entity`, `--entity-id`, `--area`, `--area-id`, `--data`, and `--return-response`.

### Automations

```bash
# List automations
hab automation list
hab automation list --extended
hab automation list --blueprint homeassistant/motion_light.yaml
hab automation list --blueprint "*"

# CRUD
hab automation create my_automation -d '{"alias":"My Automation","triggers":[],"conditions":[],"actions":[]}'
hab automation get my_automation
hab automation update my_automation -f automation.yaml
hab automation delete my_automation --force

# Run and inspect traces
hab automation run my_automation
hab automation run my_automation --skip-condition
hab automation trace my_automation

# Create from blueprint
hab automation create-from-blueprint my_motion homeassistant/motion_light.yaml -d '{"alias":"Motion","motion_entity":"binary_sensor.motion","light_target":{"entity_id":"light.kitchen"}}'
```

#### Automation triggers, conditions, and actions

```bash
# Triggers
hab automation trigger list my_automation
hab automation trigger create my_automation -d '{"trigger":"state","entity_id":"sun.sun"}'
hab automation trigger get my_automation 0
hab automation trigger update my_automation 0 -d '{"trigger":"state","entity_id":"sun.sun","to":"above_horizon"}'
hab automation trigger delete my_automation 0 --force

# Conditions
hab automation condition list my_automation
hab automation condition create my_automation -d '{"condition":"state","entity_id":"sun.sun","state":"above_horizon"}'
hab automation condition get my_automation 0
hab automation condition update my_automation 0 -d '{"condition":"state","entity_id":"sun.sun","state":"below_horizon"}'
hab automation condition delete my_automation 0 --force

# Actions
hab automation action list my_automation
hab automation action create my_automation -d '{"action":"light.turn_on","target":{"entity_id":"light.kitchen"}}'
hab automation action get my_automation 0
hab automation action update my_automation 0 -d '{"action":"light.turn_off","target":{"entity_id":"light.kitchen"}}'
hab automation action delete my_automation 0 --force
```

### Scripts

```bash
# List scripts
hab script list
hab script list --count
hab script list --brief

# CRUD and execution
hab script create evening_routine -d '{"alias":"Evening Routine","sequence":[]}'
hab script get evening_routine
hab script update evening_routine -f script.yaml
hab script run evening_routine
hab script run evening_routine -d '{"target_room":"living_room"}'
hab script delete evening_routine --force
```

#### Script actions

```bash
hab script action list evening_routine
hab script action create evening_routine -d '{"action":"light.turn_on","target":{"entity_id":"light.kitchen"}}'
hab script action get evening_routine 0
hab script action update evening_routine 0 -d '{"action":"light.turn_off","target":{"entity_id":"light.kitchen"}}'
hab script action delete evening_routine 0 --force
```

### Scenes

```bash
hab scene list
hab scene get scene.movie_mode
hab scene create movie_mode -d '{"name":"Movie Mode","entities":{"light.living_room":{"state":"on","brightness":50}}}'
hab scene update movie_mode -d '{"name":"Movie Mode Updated","entities":{}}'
hab scene activate scene.movie_mode
hab scene delete movie_mode --force
```

### Templates

```bash
# Render inline
hab template render "{{ states('sun.sun') }}"

# Render from file
hab template render -f template.j2

# Render from stdin
echo "{{ 1 + 1 }}" | hab template render
```

### Categories

```bash
hab category list --scope automation
hab category create "Security" --scope automation
hab category update <category_id> --scope automation --name "Safety"
hab category assign <category_id> automation.my_automation
hab category remove automation.my_automation
hab category delete <category_id> --scope automation --force
```

### Blueprints

```bash
# List blueprints
hab blueprint list
hab blueprint list automation
hab blueprint list script

# Import, inspect, and delete
hab blueprint import https://raw.githubusercontent.com/home-assistant/core/dev/homeassistant/components/automation/blueprints/motion_light.yaml
hab blueprint get homeassistant/motion_light.yaml
hab blueprint get --domain script my_namespace/my_script_blueprint.yaml
hab blueprint delete homeassistant/motion_light.yaml --force
```

## Dashboards

If you are new to Lovelace editing, start with:

```bash
hab guide dashboard
hab dashboard guide
```

### Dashboard CRUD

```bash
hab dashboard list
hab dashboard create my-dashboard --title "My Dashboard"
hab dashboard get my-dashboard
hab dashboard update <dashboard_id> --title "Updated Dashboard"
hab dashboard delete <dashboard_id> --force

# Read and replace raw dashboard config
hab dashboard save-config my-dashboard -f dashboard.yaml
hab dashboard get my-dashboard --json
```

### Views

```bash
hab dashboard view list my-dashboard
hab dashboard view get my-dashboard 0
hab dashboard view create my-dashboard --title "Lights" --icon mdi:lightbulb
hab dashboard view update my-dashboard 0 --title "All Lights"
hab dashboard view delete my-dashboard 0 --force
```

### Badges

```bash
hab dashboard badge list my-dashboard 0
hab dashboard badge get my-dashboard 0 0
hab dashboard badge create my-dashboard 0 --entity sun.sun
hab dashboard badge update my-dashboard 0 0 --entity person.jane_doe
hab dashboard badge delete my-dashboard 0 0 --force
```

### Sections

```bash
hab dashboard section list my-dashboard 0
hab dashboard section get my-dashboard 0 0
hab dashboard section create my-dashboard 0 --title "Climate" --type grid
hab dashboard section update my-dashboard 0 0 --title "Indoor Climate"
hab dashboard section delete my-dashboard 0 0 --force
```

### Cards

```bash
# Explicit section placement
hab dashboard card list my-dashboard 0 --section 0
hab dashboard card get my-dashboard 0 0 --section 0
hab dashboard card create my-dashboard 0 --section 0 --entity sensor.temperature
hab dashboard card update my-dashboard 0 0 --section 0 -d '{"type":"markdown","content":"Updated content"}'
hab dashboard card delete my-dashboard 0 0 --section 0 --force

# Auto-scaffold behavior: omit view/section and hab uses or creates the last one
hab dashboard card create my-dashboard --entity light.kitchen
hab dashboard card create my-dashboard --entity sun.sun --name "Sun Card"
```

`hab dashboard card create` can infer the last view and last section. If needed, it can also create missing scaffolding automatically.

## Helpers

Use `hab helper types` to discover helper families and their create parameters:

```bash
hab helper types
hab helper types --json
```

### Standard helper families

These helper families support `list`, `create`, and `delete` under `hab helper <type>`:

| Helper family | Example create command |
|---------|-------------|
| `input-boolean` | `hab helper input-boolean create "Presence" --icon mdi:toggle-switch` |
| `input-number` | `hab helper input-number create "Brightness" --min 0 --max 100 --step 5 --unit "%"` |
| `input-text` | `hab helper input-text create "Room Name" --max 50` |
| `input-select` | `hab helper input-select create "Mode" --options Home,Away,Night` |
| `input-datetime` | `hab helper input-datetime create "Wake Time" --has-time` |
| `input-button` | `hab helper input-button create "Doorbell" --icon mdi:button-pointer` |
| `counter` | `hab helper counter create "Page Views" --initial 0 --step 1 --minimum 0` |
| `timer` | `hab helper timer create "Laundry" --duration 00:45:00` |
| `schedule` | `hab helper schedule create "Office Hours"` |

Examples:

```bash
hab helper list
hab helper input-boolean list
hab helper input-boolean create "Presence"
hab helper input-boolean delete input_boolean.presence
```

### Config-flow helper families

These helpers also use `list`, `create`, and `delete`, but are backed by Home Assistant config entries:

| Helper family | Example create command |
|---------|-------------|
| `derivative` | `hab helper derivative create "Power Rate" --source sensor.power --unit-time h --round 2` |
| `integration` | `hab helper integration create "Total Energy" --source sensor.power --unit-time h --method trapezoidal` |
| `min-max` | `hab helper min-max create "Average Temp" --entities sensor.t1,sensor.t2 --type mean --round 2` |
| `threshold` | `hab helper threshold create "Freeze Alert" --entity sensor.temperature --lower 0 --hysteresis 1` |
| `utility-meter` | `hab helper utility-meter create "Monthly Energy" --source sensor.total_energy --cycle monthly` |
| `statistics` | `hab helper statistics create "Temp Average" --entity sensor.temperature --characteristic mean --sampling-size 100` |
| `local-calendar` | `hab helper local-calendar create "Work Calendar"` |
| `local-todo` | `hab helper local-todo create "Shopping"` |
| `group` | `hab helper group create "Kitchen Sensors" --type sensor --entities sensor.temp,sensor.humidity` |
| `template` | `hab helper template create "Temp Proxy" --type sensor --state "{{ 42 }}" --unit "C"` |

Examples:

```bash
hab helper derivative list
hab helper derivative create "Power Rate" --source sensor.power --unit-time h --round 2
hab helper derivative delete <entry_id>

hab helper group create "Sensor Group" --type sensor --entities sensor.one,sensor.two
hab helper group delete <entry_id>

hab helper local-calendar create "Family Calendar"
hab helper local-calendar delete <entry_id>

hab helper local-todo create "Errands"
hab helper local-todo delete <entry_id>
```

### Template helpers

Template helpers let you create helper entities backed by Jinja templates:

```bash
hab helper template list
hab helper template create "Test Sensor" --type sensor --state "{{ 42 }}" --unit "C" --device-class temperature
hab helper template create "Test Binary Sensor" --type binary_sensor --state "{{ true }}"
hab helper template create "Test Button" --type button --press homeassistant.check_config
hab helper template create "Test Number" --type number --state "{{ 50 }}" --min 0 --max 100 --step 5 --set-value input_number.set_value
hab helper template create "Test Select" --type select --state "{{ 'option1' }}" --options option1,option2 --select-option input_select.select_option
hab helper template create "Test Switch" --type switch --state "{{ false }}"
hab helper template delete <entry_id>
```

## Calendar, To-Do Lists, And Notifications

### Calendar events

```bash
# List upcoming events (defaults to the next 7 days)
hab calendar list calendar.personal
hab calendar list calendar.personal --start "2024-06-01T00:00:00Z" --end "2024-06-30T23:59:59Z"

# Create a timed event
hab calendar create calendar.personal --summary "Team Meeting" --start "2024-06-15T10:00:00" --end "2024-06-15T11:00:00" --description "Weekly sync"

# Create an all-day event
hab calendar create calendar.personal --summary "Holiday" --start "2024-12-25" --end "2024-12-26" --all-day

# Delete an event
hab calendar delete calendar.personal <event_uid>
```

### To-do lists and items

```bash
# List to-do entities
hab todo lists

# List items in a to-do entity
hab todo items todo.shopping_list

# Add items
hab todo add todo.shopping_list "Buy milk"
hab todo add todo.shopping_list "Doctor appointment" --due "2024-06-15"
hab todo add todo.shopping_list "Meeting" --due "2024-06-15T14:00:00" --description "Project review"

# Complete and uncomplete
hab todo complete todo.shopping_list abc123
hab todo uncomplete todo.shopping_list abc123

# Update and remove
hab todo update todo.shopping_list abc123 --summary "Buy oat milk" --due "2024-06-16"
hab todo remove todo.shopping_list abc123 --force
```

If you need to create a local calendar or local to-do entity first, use `hab helper local-calendar` and `hab helper local-todo`.

### Notifications

```bash
hab notification list
hab notification create "Backup completed successfully" --title "Backup"
hab notification create "Update available" --title "System" --notification-id update_notice
hab notification dismiss <notification_id>
```

`--notification-id` lets you create stable notifications that can later be updated or dismissed by ID.

## Operations And Maintenance

### Integrations

```bash
hab integration list
hab integration list --domain mqtt
hab integration get <entry_id>
hab integration reload <entry_id>
hab integration enable <entry_id>
hab integration disable <entry_id>
```

### Events

```bash
hab event list
hab event fire my_custom_event
hab event fire my_custom_event --data '{"device_id": "abc123", "action": "triggered"}'
hab event fire my_custom_event --file event_data.yaml
```

### Repairs

```bash
hab repairs list
hab repairs list --severity critical
hab repairs ignore <domain> <issue_id>
hab repairs unignore <domain> <issue_id>
```

### Backups

```bash
# Inspect backup support and configuration
hab backup agents
hab backup config get
hab backup config update --data '{"retention":{"days":7}}'

# Manage backups
hab backup list
hab backup create
hab backup create "Nightly Backup"
hab backup get <backup_id>
hab backup restore <backup_id> --agent backup.local --force
hab backup delete <backup_id> --force
```

`hab backup restore` also supports `--password`, `--restore-addon`, `--restore-folder`, `--restore-database`, and `--restore-homeassistant`.

### Energy dashboard

```bash
hab energy info
hab energy prefs get
hab energy prefs set --data '{"device_consumption":[],"device_consumption_water":[],"energy_sources":[]}'
hab energy validate
hab energy solar-forecast
```

### Diagnostics

```bash
hab diagnostics list
hab diagnostics get <domain>
```

### Network

```bash
hab network get
hab network url

# Preview network changes first
hab network configure --adapters eth0,wlan0 --plan --json

# Applying changes requires --apply and typically --force in automation
hab network configure --adapters eth0,wlan0 --apply --force
```

`hab network configure` is intentionally guarded because it can disrupt connectivity.

### Thread datasets

```bash
hab thread list
hab thread add --tlv <dataset_tlv>
hab thread get <dataset_id>
hab thread set-preferred <dataset_id>
hab thread delete <dataset_id> --plan --json
hab thread delete <dataset_id> --force
```

### System

```bash
hab system info
hab system health
hab system config-check
hab system logs
hab system updates

# Preview or execute a restart
hab system restart --plan --json
hab system restart --force
```

### Update and version

```bash
hab version
hab update --check
hab update
hab update --force
```

Automatic update checks run only for interactive sessions, are cached, and can be disabled with `--skip-update-check` or `HAB_SKIP_UPDATE_CHECK=1`.

`hab update` treats update failures as informational and does not intentionally fail the process with a non-zero exit code.

## ESPHome

Requires the ESPHome add-on or a reachable ESPHome dashboard. The ESPHome dashboard URL is auto-discovered when possible; set `HAB_ESPHOME_URL` to override.

Serial recovery commands use `esptool`; install it, set `HAB_ESPTOOL_BIN`, or use `uvx esptool`.

### Create, import, and catalog

```bash
# Create from scratch, preset, or catalog
hab esphome create living-room --platform esp32 --board nodemcu-32s --ssid MyWifi --psk secret
hab esphome create --list-presets
hab esphome create --preset-help relay
hab esphome create garage-relay --platform esp32 --board esp32dev --preset relay --relay-pin 23
hab esphome create office-plug --catalog Athom-Smart-Plug-PG01V3-EU16A

# Import a package-backed device
hab esphome import smart-plug --project-name esphome.demo --package-url https://example.com/device.yaml

# Browse and search the community device catalog
hab esphome catalog search atom --limit 5
hab esphome catalog search plug --board esp32 --type relay --difficulty 2
hab esphome catalog show M5Stack-AtomS3-Lite --include-yaml

# Discover supported boards
hab esphome boards esp32
```

### Context and configuration resolution

Many ESPHome commands resolve the target configuration from:
1. A positional argument
2. `--device`
3. Saved context from `hab esphome context use`

```bash
hab esphome context use living-room.yaml
hab esphome context show
hab esphome context clear
```

### Inspect and edit configs

```bash
hab esphome list
hab esphome info living-room.yaml
hab esphome config-read living-room.yaml
hab esphome config-write living-room.yaml -f config.yaml
hab esphome config-patch living-room.yaml --set wifi.ssid='!secret wifi_ssid'
```

### Validate, build, upload, run, and logs

```bash
hab esphome validate living-room.yaml
hab esphome validate living-room.yaml --structured --json
hab esphome build living-room.yaml
hab esphome upload living-room.yaml
hab esphome run living-room.yaml
hab esphome logs living-room.yaml
```

### Update workflow

```bash
hab esphome update living-room.yaml --set logger.level=DEBUG --upload
hab esphome update living-room.yaml --file new-config.yaml --rollback-on-fail
hab esphome update living-room.yaml --set wifi.ssid=updated-ssid --build=false
```

### Serial recovery and migration

```bash
hab esphome serial ports
hab esphome serial probe --port /dev/ttyUSB0 --chip auto
hab esphome serial erase-flash --port /dev/ttyUSB0 --chip esp32 --force

hab esphome migrate tasmota-template analyze --file template.json
hab esphome migrate tasmota-template create migrated-node --platform esp8266 --board esp01_1m --file template.json
```

### ESPHome output behavior

- `build`, `validate`, `upload`, `run`, and `logs` stream output in real time.
- In text mode, streamed lines are printed directly.
- In JSON mode, streamed commands emit NDJSON events, not a single final JSON envelope.

## Built-in Guides

Use built-in guides for workflow-level usage patterns:

```bash
hab guide
hab guide list
hab guide auth
hab guide input-output
hab guide discovery
hab guide dashboard
hab guide operations

# Legacy alias remains supported
hab dashboard guide
```

Available workflow topics:

- `index`: top-level model for using hab from an agent
- `auth`: login, status checks, and credential recovery
- `input-output`: payload strategy (`-d`, `-f`, heredoc) and JSON output usage
- `discovery`: inspect first, then mutate
- `registry`: IDs and relationships across areas, devices, and entities
- `automation`: actions, automations, scripts, scenes, templates, categories
- `dashboard`: dashboard design and resource-level editing patterns
- `helpers`: helper type selection and lifecycle workflows
- `calendar-todo`: calendar events and to-do items
- `esphome`: config validation, build, upload, and recovery workflows
- `operations`: backups, repairs, diagnostics, system, network, and integration maintenance

Use `--json` if you need machine-readable guide output:

```bash
hab guide discovery --json
hab guide operations --json
```

Guide JSON includes workflow recipes with capability requirements, command templates, verification steps, and recovery paths.

## Schema, Capabilities, And Planning

Use `hab schema` to inspect how a command should be called and what it returns:

```bash
hab schema entity get --json
hab schema dashboard card create --json
hab schema esphome logs --json
```

Use `hab capability probe` to check whether the current Home Assistant environment supports a workflow before running it:

```bash
hab capability probe --json
```

For supported mutations, use `--plan` or `--dry-run`:

```bash
hab area create "Kitchen" --plan --json
hab dashboard card create my-dashboard --entity light.kitchen --plan --json
hab helper input-boolean create "Presence" --plan --json
hab network configure --adapters eth0 --plan --json
```

Plan responses are normal JSON envelopes and may include `verification_commands`.

## Output Format

Interactive sessions default to text. Non-interactive sessions default to JSON. Use `--json` or `--text` to override.

### Success envelope

```json
{
  "success": true,
  "operation": "get",
  "resource_type": "entity",
  "data": {"entity_id": "light.living_room"},
  "message": "",
  "partial_result": false,
  "warnings": [],
  "fallbacks_applied": [],
  "missing_sections": [],
  "verification_commands": [],
  "next_suggested_commands": [],
  "error": null,
  "metadata": {
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Error envelope

```json
{
  "success": false,
  "operation": "get",
  "resource_type": "entity",
  "data": null,
  "message": "",
  "partial_result": false,
  "warnings": [],
  "fallbacks_applied": [],
  "missing_sections": [],
  "verification_commands": [],
  "next_suggested_commands": [],
  "error": {
    "code": "AUTH_REQUIRED",
    "message": "Not authenticated. Run 'hab auth login' to authenticate.",
    "details": {
      "category": "authentication",
      "retryable": false,
      "suggested_fix": "Authenticate before running commands that require API access."
    }
  },
  "metadata": {
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Metadata and partial results

When a command degrades gracefully, omits requested data, or has to resolve runtime choices, the JSON envelope can include:

- `partial_result`
- `warnings`
- `fallbacks_applied`
- `missing_sections`
- `verification_commands`
- `next_suggested_commands`
- `metadata.output_mode`
- `metadata.interactive_input`
- `metadata.interactive_output`
- `metadata.transports_used`
- `metadata.auth_source`
- `metadata.resolved_inputs`

See `LLM_CONTRACTS.md` for the machine-facing contract definitions used by `hab schema`, guide recipes, and JSON envelopes.

## Input Formats

Commands that accept data support both JSON and YAML input. Format is auto-detected from the file extension or content.

### Input methods

| Method | Flag | Description |
|--------|------|-------------|
| File | `-f`, `--file` | Read from a file (`.yaml`, `.yml`, or `.json`) |
| Inline | `-d`, `--data` | Pass data as a string argument |
| Stdin | none | Pipe data or use heredocs |

### Multi-line YAML with heredocs

```bash
hab automation create my-automation <<'EOF'
alias: Motion Light
trigger:
  - platform: state
    entity_id: binary_sensor.motion
    to: "on"
action:
  - service: light.turn_on
    target:
      entity_id: light.living_room
EOF
```

The quoted `<<'EOF'` form preserves whitespace and disables shell expansion.

### File input

```bash
hab automation create my-automation -f automation.yaml
hab dashboard view create my-dashboard -f view.yaml
hab backup config update -f backup-config.yaml
```

### Inline YAML or JSON

```bash
hab automation create test -d $'alias: Test\ntrigger:\n  - platform: state\n    entity_id: sensor.test'
hab action call light.turn_on -d '{"entity_id":"light.kitchen","brightness":200}'
```

## Features

- Hierarchical help: top-level `--help` shows command groups instead of every subcommand
- Adaptive output mode: interactive sessions default to text; non-interactive sessions default to JSON
- Structured JSON envelopes with warnings, partial-result signaling, verification commands, and structured errors
- Schema introspection via `hab schema`
- Capability probing via `hab capability probe --json`
- Mutation planning via `--plan` and `--dry-run`
- OAuth support plus token, Supervisor, and stored credential flows
- REST, WebSocket, and ESPHome dashboard transport support
- Built-in guide recipes for LLM-driven workflows
- ESPHome workflows for create, import, validate, build, upload, logs, catalog browsing, config patching, context, serial recovery, and Tasmota migration
- Automatic update checks in interactive use and self-update via `hab update`

## Commands

| Command | Description |
|---------|-------------|
| `auth` | Authentication management, discovery, refresh, and logout |
| `automation` | Manage automations, triggers, conditions, actions, runs, traces, and blueprint-backed creation |
| `script` | Manage scripts, execution, and script actions |
| `scene` | Manage scenes |
| `entity` | Entity listing, inspection, history, logbook, search, rename, enable, and disable |
| `action` | List, inspect, and call Home Assistant actions (services) |
| `area` | Manage areas |
| `floor` | Manage floors |
| `zone` | Manage zones |
| `label` | Manage labels and label assignments |
| `person` | Manage persons |
| `category` | Manage categories and category assignments |
| `helper` | Manage helper entities, config-flow helpers, local calendars, local to-dos, groups, and template helpers |
| `template` | Render Jinja templates |
| `todo` | Manage to-do list items |
| `notification` | Manage persistent notifications |
| `integration` | Manage integrations (config entries) |
| `event` | List event types and fire events |
| `repairs` | Manage Home Assistant repair issues |
| `dashboard` | Manage dashboards, views, badges, sections, cards, and raw config |
| `backup` | Inspect backup support and manage backup lifecycle |
| `energy` | Manage energy dashboard preferences and metadata |
| `diagnostics` | Inspect diagnostics handlers |
| `network` | Inspect and configure network settings |
| `calendar` | Manage calendar events |
| `capability` | Probe runtime capabilities |
| `blueprint` | Manage automation and script blueprints |
| `schema` | Show machine-readable command and output contracts |
| `system` | System information, health, restart, logs, updates, and config checks |
| `device` | Device management |
| `thread` | Manage Thread datasets and preferred network selection |
| `esphome` | Manage ESPHome devices and configs |
| `overview` | Show a high-level overview of the Home Assistant instance |
| `guide` | Display built-in usage guides |
| `search` | Search for related items and relationships |
| `update` | Update hab to the latest version |
| `version` | Show version information |

Run `hab <command> --help` for more information on a specific command group.

## Configuration

By default, configuration lives under `~/.config/home-assistant-builder/`.

- `config.json`: general settings
- `credentials.json`: encrypted credentials

If `XDG_CONFIG_HOME` is set, `hab` uses that base directory instead of `~/.config`.

Use `--config <dir>` to point `hab` at a different config directory.

### Environment variables

- `HAB_URL`: Home Assistant URL
- `HAB_TOKEN`: Home Assistant long-lived access token
- `HAB_REFRESH_TOKEN`: OAuth refresh token
- `HAB_SKIP_UPDATE_CHECK`: disable automatic update checks
- `HAB_ESPHOME_URL`: ESPHome Dashboard URL
- `HAB_ESPHOME_TOKEN`: Bearer token for ESPHome
- `HAB_ESPHOME_SESSION`: ingress session token for ESPHome access through Home Assistant
- `HAB_ESPTOOL_BIN`: path to the `esptool` binary used by ESPHome serial commands
- `HA_ACCESS_TOKEN`: alternate token source for ESPHome commands
- `SUPERVISOR_TOKEN`: Home Assistant Supervisor token fallback when running inside an add-on or app

## Development

```bash
# Clone the repository
git clone https://github.com/balloob/home-assistant-build-cli
cd home-assistant-build-cli

# Build
go build -o hab .

# Run unit tests
go test ./...

# Run integration tests (requires empty-hass)
./test/run_integration_test.sh
```

## License

Apache 2.0 License.

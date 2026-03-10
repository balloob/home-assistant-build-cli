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

### Authentication

```bash
# Authenticate using OAuth
hab auth login

# Or use a long-lived access token
hab auth login --token --url http://homeassistant.local:8123 --access-token "your_token"

# Check authentication status
hab auth status
```

### Basic Commands

```bash
# List entities
hab entity list
hab entity list --domain light

# Get entity state
hab entity get light.living_room

# Call actions
hab action call light.turn_on --entity light.living_room --data '{"brightness": 200}'

# List automations
hab automation list

# Manage areas
hab area list
hab area create "Kitchen"
```

### Scenes

```bash
# List all scenes
hab scene list

# Get scene details
hab scene get scene.movie_mode

# Create a scene
hab scene create movie_mode -d '{"alias": "Movie Mode", "entities": {"light.living_room": {"state": "on", "brightness": 50}}}'

# Activate a scene
hab scene activate scene.movie_mode

# Delete a scene
hab scene delete movie_mode
```

### Persons

```bash
# List all persons
hab person list

# Create a person
hab person create "John Doe"

# Update a person
hab person update <person_id> --name "Jane Doe"

# Delete a person
hab person delete <person_id>
```

### Categories

```bash
# List categories for a scope
hab category list automation

# Create a category
hab category create automation "Security"

# Assign a category to an entity
hab category assign automation <entity_id> <category_id>

# Remove a category assignment
hab category remove automation <entity_id>
```

### Templates

```bash
# Render a Jinja2 template inline
hab template render "{{ states('sun.sun') }}"

# Render from a file
hab template render -f template.j2
```

### Entity Logbook

```bash
# View logbook entries for an entity
hab entity logbook light.living_room
hab entity logbook light.living_room --start "2024-01-01T00:00:00Z" --end "2024-01-02T00:00:00Z"
```

### To-Do Lists

```bash
# List all to-do list entities
hab todo lists

# List items in a to-do list
hab todo items todo.shopping_list

# Add an item
hab todo add todo.shopping_list "Buy milk"
hab todo add todo.shopping_list "Doctor appointment" --due "2024-06-15"
hab todo add todo.shopping_list "Meeting" --due "2024-06-15T14:00:00" --description "Project review"

# Complete / uncomplete an item
hab todo complete todo.shopping_list "Buy milk"
hab todo uncomplete todo.shopping_list "Buy milk"

# Update an item
hab todo update todo.shopping_list "Buy milk" --summary "Buy oat milk" --due "2024-06-16"

# Remove an item
hab todo remove todo.shopping_list "Buy milk"
```

### Notifications

```bash
# List persistent notifications
hab notification list

# Create a notification
hab notification create "Backup completed successfully" --title "Backup"

# Dismiss a notification
hab notification dismiss <notification_id>
```

### Calendar Events

```bash
# List upcoming events (defaults to next 7 days)
hab calendar list calendar.personal
hab calendar list calendar.personal --start "2024-06-01T00:00:00Z" --end "2024-06-30T23:59:59Z"

# Create a timed event
hab calendar create calendar.personal "Team Meeting" --start "2024-06-15T10:00:00" --end "2024-06-15T11:00:00"

# Create an all-day event
hab calendar create calendar.personal "Holiday" --start "2024-12-25" --end "2024-12-26" --all-day

# Delete an event
hab calendar delete calendar.personal <event_uid>
```

### Integrations

```bash
# List all integrations
hab integration list
hab integration list --domain mqtt

# Get details for a specific integration
hab integration get <entry_id>

# Reload an integration
hab integration reload <entry_id>

# Enable / disable an integration
hab integration enable <entry_id>
hab integration disable <entry_id>
```

### Events

```bash
# List all registered event types
hab event list

# Fire a custom event
hab event fire my_custom_event
hab event fire my_custom_event --data '{"device_id": "abc123", "action": "triggered"}'
hab event fire my_custom_event --file event_data.yaml
```

### Repairs

```bash
# List all repair issues
hab repairs list
hab repairs list --severity critical

# Ignore / unignore a repair issue
hab repairs ignore <domain> <issue_id>
hab repairs unignore <domain> <issue_id>
```

### ESPHome

Requires the ESPHome add-on. The ESPHome Dashboard URL is auto-discovered via the HA Supervisor; set `HAB_ESPHOME_URL` to override.

```bash
# List devices and their status
hab esphome list

# Read/write device configs
hab esphome config-read living-room.yaml
hab esphome config-write living-room.yaml -f config.yaml

# Validate, build, and flash
hab esphome validate living-room.yaml
hab esphome build living-room.yaml
hab esphome upload living-room.yaml
hab esphome run living-room.yaml

# Stream live logs
hab esphome logs living-room.yaml
```

Some ESPHome commands (`build`, `validate`, `run`, `upload`, `logs`) stream output in real-time rather than returning a single JSON envelope.

## Features

- **Hierarchical Help**: Top-level `--help` shows command groups, not all sub-commands
- **Text Output**: Human-readable text output by default
- **JSON Mode**: Machine-parseable JSON with `--json` flag
- **OAuth Support**: Full OAuth2 flow for authentication
- **WebSocket & REST**: Uses both APIs for optimal functionality
- **Auto-Update**: Checks for updates automatically and supports self-updating via `hab update`

## Commands

| Command | Description |
|---------|-------------|
| `auth` | Authentication management |
| `automation` | Manage automations |
| `script` | Manage scripts |
| `scene` | Manage scenes |
| `entity` | Entity operations (includes `logbook` subcommand) |
| `action` | Call actions |
| `area` | Manage areas |
| `floor` | Manage floors |
| `zone` | Manage zones |
| `label` | Manage labels |
| `person` | Manage persons |
| `category` | Manage entity categories |
| `helper` | Manage helper entities |
| `template` | Render Jinja2 templates |
| `todo` | Manage to-do list items |
| `notification` | Manage persistent notifications |
| `integration` | Manage integrations (config entries) |
| `event` | List event types and fire events |
| `repairs` | Manage Home Assistant repair issues |
| `dashboard` | Manage dashboards |
| `backup` | Backup and restore |
| `calendar` | Manage calendar events (includes `create` and `delete` subcommands) |
| `blueprint` | Manage blueprints |
| `system` | System operations |
| `device` | Device management |
| `thread` | Manage Thread credentials |
| `esphome` | Manage ESPHome devices |
| `overview` | Show an overview of the HA instance |
| `search` | Search for items and relationships |
| `update` | Update hab to the latest version |
| `version` | Show version information |

Run `hab <command> --help` for more information on each command.

## Output Format

By default, all commands output human-readable text. Use `--json` for machine-parseable JSON:

```bash
hab entity get light.living_room --json
```

JSON output uses a standard envelope:

```json
{
  "success": true,
  "data": { ... },
  "metadata": {
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Input Formats

Commands that accept data (automations, dashboards, scripts, etc.) support both **JSON** and **YAML** input. The format is auto-detected based on file extension or content structure.

### Input Methods

| Method | Flag | Description |
|--------|------|-------------|
| File | `-f`, `--file` | Read from a file (`.yaml`, `.yml`, or `.json`) |
| Inline | `-d`, `--data` | Pass data as a string argument |
| Stdin | (none) | Pipe data or use heredocs |

### Multi-line YAML with Heredocs

For multi-line YAML where whitespace matters, use a heredoc:

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

The `<<'EOF'` syntax (with quotes) preserves exact whitespace and prevents shell variable expansion.

### File Input

```bash
hab automation create my-automation -f automation.yaml
hab dashboard view create my-dashboard -f view.yaml
```

### Inline YAML (short configs)

Use `$'...'` syntax for short inline YAML with newlines:

```bash
hab automation create test -d $'alias: Test\ntrigger:\n  - platform: state\n    entity_id: sensor.test'
```

## Configuration

Configuration is stored in `~/.config/home-assistant-builder/`:

- `config.json` - General settings
- `credentials.json` - Encrypted credentials

### Environment Variables

- `HAB_URL` - Home Assistant URL
- `HAB_TOKEN` - Long-lived access token
- `HAB_CONFIG_DIR` - Custom config directory
- `HAB_ESPHOME_URL` - ESPHome Dashboard URL (auto-discovered from HA Supervisor if not set)
- `HAB_ESPHOME_TOKEN` - Bearer token for ESPHome ingress proxy (overrides default credentials)
- `HAB_ESPHOME_SESSION` - Ingress session token for ESPHome (required when accessing via HA Core ingress)

## Development

```bash
# Clone the repository
git clone https://github.com/balloob/home-assistant-build-cli
cd home-assistant-build-cli

# Build
go build -o hab .

# Run tests
go test ./...

# Run integration tests (requires empty-hass)
./test/run_integration_test.sh
```

## License

Apache 2.0 License.

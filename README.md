# Home Assistant Builder (hab)

A CLI utility designed for LLMs to build and manage Home Assistant configurations.

_Vibe coded, use at own risk._

## Installation

### From Source

```bash
go install github.com/home-assistant/hab@latest
```

Or build locally:

```bash
git clone https://github.com/home-assistant/hab
cd hab
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

## Features

- **Hierarchical Help**: Top-level `--help` shows command groups, not all sub-commands
- **Structured Output**: JSON output by default for easy parsing
- **Text Mode**: Human-readable output with `--text` flag
- **OAuth Support**: Full OAuth2 flow for authentication
- **WebSocket & REST**: Uses both APIs for optimal functionality
- **Auto-Update**: Checks for updates automatically and supports self-updating via `hab update`

## Commands

| Command | Description |
|---------|-------------|
| `auth` | Authentication management |
| `automation` | Manage automations |
| `script` | Manage scripts |
| `entity` | Entity operations |
| `action` | Call actions |
| `area` | Manage areas |
| `floor` | Manage floors |
| `zone` | Manage zones |
| `label` | Manage labels |
| `helper` | Manage helper entities |
| `dashboard` | Manage dashboards |
| `backup` | Backup and restore |
| `calendar` | Manage calendar events |
| `blueprint` | Manage blueprints |
| `system` | System operations |
| `device` | Device management |
| `group` | Manage entity groups |
| `thread` | Manage Thread credentials |
| `search` | Search for items and relationships |
| `update` | Update hab to the latest version |
| `version` | Show version information |

Run `hab <command> --help` for more information on each command.

## Output Format

By default, all commands output JSON:

```json
{
  "success": true,
  "data": { ... },
  "metadata": {
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

Use `--text` for human-readable output:

```bash
hab entity get light.living_room --text
```

## Configuration

Configuration is stored in `~/.config/home-assistant-builder/`:

- `config.json` - General settings
- `credentials.json` - Encrypted credentials

### Environment Variables

- `HAB_URL` - Home Assistant URL
- `HAB_TOKEN` - Long-lived access token
- `HAB_CONFIG_DIR` - Custom config directory

## Development

```bash
# Clone the repository
git clone https://github.com/home-assistant/hab
cd hab

# Build
go build -o hab .

# Run tests
go test ./...

# Run integration tests (requires empty-hass)
./test/run_integration_test.sh
```

## License

Apache 2.0 License.

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Home Assistant Builder (`hab`) is a CLI utility designed for LLMs to build and manage Home Assistant configurations. It outputs JSON by default for easy parsing and uses both REST and WebSocket APIs to communicate with Home Assistant.

## Build and Test Commands

```bash
# Build
go build -o hab .

# Run unit tests
go test ./...

# Run integration tests (requires uvx and empty-hass)
./test/run_integration_test.sh
```

## Architecture

### Package Structure

- **cmd/**: Cobra command definitions organized by feature (auth, entity, automation, etc.)
- **auth/**: Authentication handling - OAuth flow, token refresh, credential storage
- **client/**: API clients (RestClient for HTTP, output formatting)
- **config/**: Configuration paths and settings (uses viper)
- **input/**: Input parsing for YAML/JSON data

### Key Patterns

**Command Structure**: Each feature has a parent command file (`cmd/entity.go`) and subcommand files (`cmd/entity_list.go`, `cmd/entity_get.go`, etc.). The parent registers subcommands and the root command.

**Output Format**: All commands use `client.PrintSuccess()` or `client.FormatError()` for consistent JSON output. Text mode (`--text`) uses `client.FormatOutput()` with `textMode=true`.

**Authentication Flow**: Commands obtain a configured REST client via `auth.Manager.GetRestClient()`, which handles credential loading and automatic token refresh.

**Configuration**: Uses viper for config management with environment variable prefix `HAB_` (e.g., `HAB_URL`, `HAB_TOKEN`). Config stored in `~/.config/home-assistant-builder/`.

### API Communication

- REST API via `client.RestClient` (uses resty) for state queries, service calls
- WebSocket API via `client.WebSocketClient` for registry operations (areas, floors, labels, devices)

### Learning Domain Interactions

To understand how to interact with specific Home Assistant domains (e.g., `light`, `climate`, `cover`), check the data folder of the Home Assistant frontend repository:

- **Web**: https://github.com/home-assistant/frontend/tree/dev/src/data
- **CLI**: `gh browse home-assistant/frontend:src/data` or `gh api repos/home-assistant/frontend/contents/src/data`

Each domain typically has a TypeScript file (e.g., `light.ts`, `climate.ts`) that defines the available services, attributes, and WebSocket commands.

### Adding New Commands

When adding new commands:
1. Follow the existing command structure pattern (parent command + subcommand files)
2. **Always add tests** for new commands in `test/run_integration_test.sh`
3. Use `client.PrintOutput()` or `client.PrintSuccess()` for consistent output

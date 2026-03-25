# AGENTS.md

## Response Behavior

- When a task is complete, state what was done and stop. Do NOT suggest next steps, follow-up improvements, or additional actions unless explicitly asked.
- Do not end responses with questions like "Would you like me to..." or "Should I also...".
- Do not list potential improvements, optimizations, or refactoring opportunities unless requested.
- If something is ambiguous, ask a single clarifying question before proceeding rather than guessing and then suggesting alternatives after the fact.

## Code Quality Principles

### General

- Apply DRY (Don't Repeat Yourself): extract shared logic into helper functions or shared types rather than duplicating code across commands or packages.
- Apply SOLID principles where idiomatic to Go: single-responsibility functions and packages, dependency injection via interfaces, and open/closed design through composition rather than inheritance.
- Apply Separation of Concerns: keep command definitions, business logic, API communication, and output formatting in their respective packages. Do not mix CLI flag parsing with API call logic.
- Before writing new utility functions, check whether the standard library or an existing project helper already provides the functionality.
- Prefer small, focused functions. If a function exceeds ~40 lines, consider whether it can be decomposed.
- Handle all errors explicitly. Never discard errors with `_` unless there is a clear, commented reason.

### Modern Go (1.24+)

LLMs are known to produce outdated Go idioms. Always prefer modern Go constructs:

- **`range` over integers**: use `for i := range n` instead of `for i := 0; i < n; i++` (Go 1.22+).
- **`range` over functions / iterators**: use `iter.Seq` and `iter.Seq2` with `for v := range collection.All()` instead of manual `.Next()` / `.HasNext()` patterns (Go 1.23+).
- **`slices` and `maps` packages**: use `slices.Contains`, `slices.SortFunc`, `maps.Keys`, `maps.Clone` etc. instead of hand-rolled loops for common collection operations (Go 1.21+).
- **`errors.AsType`**: use the generic `errors.AsType[*MyError](err)` instead of the older `errors.As(err, &target)` pattern. It is type-safe, faster, and scopes variables to their `if` block (Go 1.26+).
- **`os.Root`**: use for traversal-resistant filesystem access when handling user-supplied paths (Go 1.24+).
- **Generic type aliases**: supported as of Go 1.24. Use when they improve readability.
- **`testing/synctest`**: use for testing concurrent / async code instead of manual sleeps and polling (Go 1.25+).
- **`go fix`**: when modernizing existing code, run `go fix ./...` to automatically apply available modernizers from gopls.

When in doubt about whether a modern API exists, check the Go release notes or standard library documentation rather than defaulting to older patterns.

### Go Idioms

- Accept interfaces, return structs.
- Communicate via channels for coordination; use mutexes only for protecting shared state.
- Keep interface definitions small (1-3 methods). Define interfaces at the consumer, not the provider.
- Use `context.Context` as the first parameter for functions that perform I/O or may be cancelled.
- Wrap errors with `fmt.Errorf("...: %w", err)` to preserve the error chain.
- Use table-driven tests with `t.Run` subtests.
- Run `go vet ./...` and ensure code passes before considering a task complete.

## Project Overview

Home Assistant Builder (`hab`) is a CLI utility designed for LLMs to build and manage Home Assistant configurations. It outputs human-readable text by default and supports machine-parseable JSON with `--json`. It uses both REST and WebSocket APIs to communicate with Home Assistant.

## Build and Test Commands

```bash
# Build
go build -o hab .

# Run unit tests
go test ./...

# Run all integration tests (requires uvx and empty-hass)
./test/run_integration_test.sh

# Run specific test group
./test/run_integration_test.sh core        # Auth & system tests
./test/run_integration_test.sh registry    # Entity/device/area/floor/label tests
./test/run_integration_test.sh automation  # Automation tests
./test/run_integration_test.sh script      # Script tests
./test/run_integration_test.sh dashboard   # Dashboard tests
./test/run_integration_test.sh helpers     # Helper type tests
./test/run_integration_test.sh template    # Template entity tests
./test/run_integration_test.sh calendar    # Calendar and to-do list tests
./test/run_integration_test.sh misc        # Actions, zones, backups, etc.

# Run a single test file standalone (starts its own empty-hass)
./test/test_automation.sh
```

### Integration Test Structure

Tests are organized by feature into separate files:

- **test/lib/common.sh**: Shared functions, colors, test helpers
- **test/test_core.sh**: Auth login/logout/status, system info/health
- **test/test_registry.sh**: Entity, device, area, floor, label, person CRUD operations; entity logbook
- **test/test_automation.sh**: Automation and automation-trigger/condition/action CRUD; scene CRUD
- **test/test_script.sh**: Script and script-action CRUD
- **test/test_dashboard.sh**: Dashboard, views, badges, sections, cards CRUD
- **test/test_helpers.sh**: Helper types (input_boolean, counter, timer, group, etc.)
- **test/test_template.sh**: Template entity types (sensor, binary_sensor, switch, number, etc.)
- **test/test_calendar_todo.sh**: Local calendar and to-do list helpers; todo item CRUD; calendar create/delete
- **test/test_misc.sh**: Actions, zones, backups, blueprints, categories, template render, notifications, integrations, events, repairs

Each test file can:
1. Run **standalone**: `./test/test_automation.sh` - starts its own empty-hass instance
2. Run **via orchestrator**: `./test/run_integration_test.sh automation` - uses shared empty-hass

When running all tests via `./test/run_integration_test.sh`, empty-hass is started once and shared across all test files.

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
2. **Always add tests** for new commands in the appropriate test file under `test/`:
   - Entity/device/area/floor/label commands: `test/test_registry.sh`
   - Automation commands: `test/test_automation.sh`
   - Script commands: `test/test_script.sh`
   - Dashboard commands: `test/test_dashboard.sh`
   - Helper commands: `test/test_helpers.sh`
   - Template entity commands: `test/test_template.sh`
   - Calendar/to-do commands: `test/test_calendar_todo.sh`
   - Other commands: `test/test_misc.sh`
3. Use `client.PrintOutput()` or `client.PrintSuccess()` for consistent output

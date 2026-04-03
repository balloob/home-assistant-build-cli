# hab Guide

Use this guide to drive `hab` effectively from an LLM or other automation.

## When to Use

- Start here before any new automation workflow.
- Use this topic to choose the correct domain-specific guide.

## Prerequisites

- Authenticate first when you expect to run write commands: `hab auth login`.
- Use `--json` when results will be parsed by an agent.

## Discovery Sequence

1. `hab guide list --json`
2. `hab overview --json`
3. `hab entity search <query> --json`
4. `hab action docs <domain.action> --json`

## Mutation Pattern

- Apply one change at a time.
- Reuse discovered IDs exactly as returned by `list/get/search` commands.

## Verification Commands

```bash
hab overview --json
hab entity get <entity_id> --json
```

## Pitfalls

- Mutating resources with guessed IDs.
- Parsing text output in automation instead of JSON envelopes.

## Topic Guides

- `hab guide auth` - login, status checks, token refresh, and auth recovery.
- `hab guide input-output` - data input styles, JSON output, and payload hygiene.
- `hab guide discovery` - inspect and identify the right targets.
- `hab guide registry` - areas/floors/devices/entities/labels/persons/zones.
- `hab guide automation` - actions, automations, scripts, scenes, templates, categories.
- `hab guide dashboard` - dashboard modeling and card/view best practices.
- `hab guide helpers` - helper type discovery and helper creation patterns.
- `hab guide calendar-todo` - event and task workflows.
- `hab guide esphome` - ESPHome config/build/upload workflows.
- `hab guide operations` - backups, diagnostics, repairs, network, and system maintenance.

## JSON Examples

```bash
hab guide list --json
hab guide discovery --json
hab entity list --domain light --json
```

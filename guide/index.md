# hab Guide

Use this guide to drive `hab` effectively from an LLM or other automation.

## Operating Model

- Use `--json` for machine parsing.
- Use text mode for human-readable guidance and summaries.
- Use `hab <command> --help` for exact flags and arguments.
- Use discovery commands before making changes.

## Fast Start Workflow

1. Authenticate if needed: `hab auth login`
2. Inspect instance state: `hab overview --json`
3. Find target entities/devices: `hab entity search <query> --json` and `hab device list --json`
4. Inspect action schema: `hab action docs <domain.action> --json`
5. Execute change: `hab action call ...`
6. Verify state: `hab entity get <entity_id> --json`

## Topic Guides

- `hab guide discovery` - inspect and identify the right targets.
- `hab guide registry` - areas/floors/devices/entities/labels/persons/zones.
- `hab guide automation` - actions, automations, scripts, scenes, templates, categories.
- `hab guide dashboard` - dashboard modeling and card/view best practices.
- `hab guide helpers` - helper type discovery and helper creation patterns.

## JSON Examples

```bash
hab guide list --json
hab guide discovery --json
hab entity list --domain light --json
```

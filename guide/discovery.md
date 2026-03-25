# Discovery Workflows

Use this topic before edits. It helps you identify the right entities, devices, and actions.

## Recommended Sequence

1. Instance overview: `hab overview --json`
2. Find entities by intent: `hab entity search kitchen --json`
3. Inspect device relationships: `hab device list --json` and `hab device entities <device_id> --json`
4. Find cross-resource links: `hab search related entity light.kitchen --json`
5. Inspect action contracts: `hab action list light --json` and `hab action docs light.turn_on --json`

## Useful Discovery Commands

```bash
hab overview --json
hab entity list --domain sensor --limit 20 --json
hab device list --area living_room --json
hab search related --type area --id living_room --json
hab action data --json
```

## Notes for LLM Usage

- Prefer entity IDs from command output over guessed names.
- If a command supports both positional args and flags, keep one style consistent in a workflow.
- Verify output after each mutation step with `get` or `list` commands.

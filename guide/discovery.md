# Discovery Workflows

Use this topic before edits. It helps you identify the right entities, devices, and actions.

## When to Use

- Before any mutation command.
- When a workflow fails and you need to re-check IDs, relationships, or action fields.

## Prerequisites

- Prefer JSON mode while gathering data: `--json`.
- Use an authenticated session if discovery will be followed by writes.

## Discovery Sequence

1. Instance overview: `hab overview --json`
2. Find entities by intent: `hab entity search kitchen --json`
3. Inspect device relationships: `hab device list --json` and `hab device entities <device_id> --json`
4. Find cross-resource links: `hab search related entity light.kitchen --json`
5. Inspect action contracts: `hab action list light --json` and `hab action docs light.turn_on --json`

## Mutation Pattern

- Keep discovery and mutation as separate steps.
- Capture identifiers from output and feed them directly into write commands.

## Useful Discovery Commands

```bash
hab overview --json
hab entity list --domain sensor --limit 20 --json
hab device list --area living_room --json
hab search related --type area --id living_room --json
hab action data --json
```

## Verification Commands

```bash
hab entity get <entity_id> --json
hab device entities <device_id> --json
```

## Pitfalls

- Using guessed entity names instead of discovered IDs.
- Calling action APIs without checking docs for required payload fields.

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

## Short Resource IDs

Discovery lists preserve canonical IDs and add `short_id` for long hexadecimal,
UUID, and ULID identifiers, including integration-prefixed backup agent IDs. Copy that value into any command argument or flag
that selects the same resource. Prefixes are case-sensitive, at least eight
characters, and must identify exactly one item. Exact IDs always take priority.
Readable IDs such as `light.kitchen`, area slugs, and dashboard paths stay exact.

Abbreviations are calculated across the entire resource collection before list
filters or limits. Commands resolve them again against current Home Assistant
data before previews, confirmations, or writes. If a prefix becomes ambiguous,
the error includes full IDs and names; use a longer prefix. ULIDs created close
together can require more than eight characters because they share a timestamp.

Use the general discovery command for IDs without a dedicated list command:

```bash
hab id list device --json
hab id list user --json
hab id list config_entry --json
hab id list category --scope automation --json
hab id list todo_item --scope todo.shopping_list --json
hab id list trace --scope automation.kitchen --json
```

Scopes matter: categories belong to a category scope, to-do items to a list,
traces to an automation, repairs to an integration domain, and calendar events
to a calendar and time window. `hab id list --help` lists resource types.

Typed references such as `device_id`, `area_id`, `floor_id`, `label_id`,
`user_id`, and `config_entry_id` also accept short IDs in JSON/YAML configuration
and action data. New IDs, names, templates, and arbitrary variables are preserved.
JSON responses record expansions in `metadata.resolved_ids`. Selector prefixes
are resolved before plans; static plans still redact opaque payloads without
parsing them. Resolution requires a live authenticated connection even when
a command's usual preview is static. Command schemas expose `short_id_inputs`
with the resource, scope, minimum length, and discovery transport.

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

# Registry Commands

Use this topic for registry resources: areas, floors, labels, devices, entities, persons, and zones.

## When to Use

- When creating, updating, or deleting registry resources.
- When mapping relationships between entities, devices, and areas.

## Prerequisites

- Run discovery commands first to collect valid IDs.
- Use JSON mode for machine parsing.

## Identifier Rules

- `entity` commands use `entity_id` values like `light.kitchen`.
- `device` commands use device registry IDs.
- `area`, `floor`, `label`, `person`, and `zone` commands use their own IDs.
- Use list/get commands to discover IDs before update/delete commands.

## Area Metadata

`area update` also supports `--aliases`, `--labels`, `--picture`,
`--temperature-entity`, and `--humidity-entity`. Aliases and labels accept
comma-separated or repeated values and replace the entire set. Aliases are
trimmed and deduplicated. Empty values clear each field; omitted flags preserve
existing values. Pictures are stored as a URL/path, not uploaded by the command.

```bash
hab area update <area_id> --aliases "Kitchen,Cooking area" --labels downstairs --plan --json
hab area update <area_id> --picture /local/kitchen.jpg --json
hab area update <area_id> --temperature-entity sensor.kitchen_temperature --humidity-entity sensor.kitchen_humidity --json
hab area update <area_id> --aliases "" --labels "" --picture "" --temperature-entity "" --humidity-entity "" --json
```

Floor and label IDs are validated before writing. Temperature/humidity assignments
must reference existing sensor states with the matching device class. Preview
performs the same checks; an invalid reference prevents the entire update.

## Discovery Sequence

1. List top-level resources (`area`, `floor`, `label`).
2. List devices and map to entities.
3. Confirm IDs with `get` before mutation commands.

## Mutation Pattern

- Create resources first, then assign relationships.
- Update and delete only with explicit IDs.

## Common Commands

```bash
hab area list --json
hab area create "Kitchen"
hab floor list --json
hab label list --json
hab device list --area kitchen --json
hab entity list --device <device_id> --json
hab person list --json
hab zone create "Office" --latitude 37.7749 --longitude -122.4194 --radius 100
```

## Verification Commands

```bash
hab area get <area_id> --json
hab device get <device_id> --json
hab entity get <entity_id> --json
hab person get <person_id> --json
hab zone list --json
```

## Updating Entities

`entity update` changes the registry ID, friendly name, or icon. Provide at least
one of `--entity-id`, `--name`, or `--icon`; they can be used independently or
combined in a single verified update. `entity rename` remains available for
friendly-name changes.

```bash
hab entity update switch.plug_2 --entity-id switch.kitchen_plug --plan --json
hab entity update switch.plug_2 --entity-id switch.kitchen_plug --json
hab entity get switch.kitchen_plug --json
hab entity update switch.kitchen_plug --name "Kitchen plug" --icon mdi:power-socket-eu --plan --json
hab entity update switch.kitchen_plug --name "Kitchen plug" --icon mdi:power-socket-eu --json
hab entity update switch.kitchen_plug --name "" --icon "" --json
```

Omitted fields are preserved. Empty `--name ""` and `--icon ""` values clear the
registry overrides, restoring the default name or icon. The preview's `changes`
object shows before/after values for each requested field; `null` clears an
override. Name/icon-only changes do not need collision or relationship searches.

For ID changes, the new ID must keep the same domain. Both preview (`--plan`, also
`--dry-run`) and apply check the registry and live states for collisions before writing.
`unique_id` stays untouched, preserving the integration link.

The preview includes old/new IDs, related items from `hab search related`, and
warnings. Related items are candidates for review, not proof of broken references.
Search does not cover dashboards, templates, or every dynamic reference; inspect
those manually even if the related list is empty. No references are rewritten.

Long-term statistics use `statistic_id`. Home Assistant's recorder normally
migrates history and statistics after a registry rename, but existing history or
statistics at the target ID can prevent migration. The command does not verify
or repair this migration. Check power/energy history after renaming.

The result reports `saved` and `verified` separately. `saved: true` means Home
Assistant acknowledged the registry write; `verified: true` means a separate
read-back confirmed every requested field and unchanged `unique_id`. Errors
include the result under `error.details.result`; `saved: null` means
acknowledgement was uncertain. Inspect the registry (both IDs for an ID change)
before retrying an uncertain or unverified write.

## Pitfalls

- Passing a device ID where an entity ID is required.
- Deleting resources by name assumptions instead of confirmed IDs.

## Updating Areas, Floors, and Labels

Update commands read current registry values for `--plan` / `--dry-run` and
report `changes` with before/after values and an observed `would_change`.
Planning requires a connection to Home Assistant. Apply verifies each requested
field using a separate registry read. Existing resource fields remain in `data`,
alongside `mode`, `changes`, `would_change`, `saved`, and `verified`.

Omitted flags preserve fields. Empty optional values clear them: area `--floor`
and `--icon`; floor `--icon` and `--level`; label `--icon`, `--color`, and
`--description`. Names are required and cannot be cleared. Floor `--level 0`
sets ground level; `--level ""` removes the level.

```bash
hab area update <area_id> --floor "" --plan --json
hab floor update <floor_id> --level "" --json
hab label update <label_id> --icon "" --color "" --description "" --json
```

A no-op returns `saved: false, verified: true` without writing. A preview returns
`saved: false, verified: false`. A successful write returns both as true.
Errors include these fields under `error.details.result`. `saved: null` means
the write acknowledgement is uncertain; inspect the registry before retrying.
Verification confirms registry metadata, not runtime entity availability.

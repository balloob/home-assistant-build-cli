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

## Floor Aliases

Use `floor update --aliases` to replace the floor's alias set. Accepts
comma-separated or repeated flags; whitespace is trimmed, duplicates and empty
aliases are removed. `--aliases ""` clears all aliases, while omitting the flag
preserves existing aliases. Previews and read-back compare aliases as a set,
so a different server ordering does not produce a false mismatch.

```bash
hab floor update <floor_id> --aliases "Ground floor,Downstairs" --plan --json
hab floor update <floor_id> --aliases "Ground floor" --aliases Downstairs --json
hab floor update <floor_id> --aliases "" --json
```

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

`entity update` changes entity registry settings. Provide at least one field flag;
they can be used independently or combined in a single verified update.
`entity rename` remains available for friendly-name changes.

| Flag | Behavior |
| --- | --- |
| `--entity-id` | Change the ID within the same domain. |
| `--name`, `--icon`, `--device-class` | Set an override; an empty string clears it. Use a device class supported by the entity's domain. |
| `--area` | Set an area ID; an empty string restores area inheritance from the device. |
| `--labels` | Replace all label IDs with a comma-separated or repeated list; `--labels ""` clears them. |
| `--aliases` | Replace the ordered alias list (CSV or repeated); `--aliases ""` clears it. Whitespace is trimmed and blank aliases are removed. |
| `--category scope=id` | Set a category in a scope; repeat for multiple scopes. `--category helpers=` removes only that scope. |
| `--hidden`, `--disabled` | Set the user override; `--hidden=false` unhides and `--disabled=false` enables. |

Category scopes are `automation`, `script`, `scene`, and `helpers`. Other scopes
are preserved. Discover IDs with `area list`, `label list`, and
`category list --scope helpers`. Supplied area, label, and category IDs are
checked before writing, including during preview. Integration-specific options
are not supported.

```bash
hab entity update switch.plug_2 --entity-id switch.kitchen_plug --plan --json
hab entity update switch.plug_2 --entity-id switch.kitchen_plug --json
hab entity get switch.kitchen_plug --json
hab entity update switch.kitchen_plug --name "Kitchen plug" --icon mdi:power-socket-eu --plan --json
hab entity update switch.kitchen_plug --name "Kitchen plug" --icon mdi:power-socket-eu --json
hab entity update switch.kitchen_plug --name "" --icon "" --json
hab entity update switch.kitchen_plug --area kitchen --labels lighting,downstairs --aliases "Kitchen plug,Counter socket" --plan --json
hab entity update input_boolean.guest_mode --category helpers=guest_settings --json
hab entity update switch.kitchen_plug --hidden=false --disabled=false --json
```

Omitted fields are preserved. Empty string overrides restore defaults. The preview's `changes`
object shows before/after values for each requested field; `null` clears an
override. Labels and aliases use empty arrays to clear the list. Category changes
show the resulting map with unrelated scopes preserved. Only ID changes need
entity collision or relationship searches. Alias order is preserved; on Home
Assistant versions that use the first alias as the primary name, it matters.

Disabling removes an entity from active use. Enabling may require a reload or
restart; the result includes Home Assistant's activation warnings. An entity
whose device or integration is disabled may remain unavailable. Verification
checks registry fields, not whether the entity is running or visible in a dashboard.

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

## Updating Devices

`device update` supports `--name` (the user name override), `--area`, `--labels`,
and `--disabled`. Empty name/area values clear the override or assignment.
Labels accept comma-separated or repeated values and replace the entire set;
`--labels ""` removes all labels. Area and label IDs are checked before writing.

```bash
hab device update <device_id> --name "Kitchen plug" --area kitchen --plan --json
hab device update <device_id> --labels lighting,downstairs --json
hab device update <device_id> --name "" --area "" --labels "" --disabled=false --json
```

`--disabled` sets the user disable reason; `--disabled=false` clears it. Disabling
a device affects its entities. Read-back verifies registry metadata, not whether
those entities have finished unloading or reloading. Preview, no-op, `saved`,
and `verified` behavior matches area/floor/label updates. Integration-specific
options are not supported.

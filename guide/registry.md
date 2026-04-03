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

## Pitfalls

- Passing a device ID where an entity ID is required.
- Deleting resources by name assumptions instead of confirmed IDs.

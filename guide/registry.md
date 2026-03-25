# Registry Commands

Use this topic for registry resources: areas, floors, labels, devices, entities, persons, and zones.

## Identifier Rules

- `entity` commands use `entity_id` values like `light.kitchen`.
- `device` commands use device registry IDs.
- `area`, `floor`, `label`, `person`, and `zone` commands use their own IDs.
- Use list/get commands to discover IDs before update/delete commands.

## Common Patterns

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

## Verify After Changes

```bash
hab area get <area_id> --json
hab device get <device_id> --json
hab entity get <entity_id> --json
hab person get <person_id> --json
hab zone list --json
```

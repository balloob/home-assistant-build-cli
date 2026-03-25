# Helper Commands

Use this topic for helper entities such as input booleans, counters, timers, groups, and template helpers.

## Discover Helper Types and Parameters

```bash
hab helper types --json
hab helper input-boolean --help
hab helper statistics create --help
```

## Common Helper Workflows

```bash
hab helper input-boolean create "Guest Mode" --icon mdi:account
hab helper input-boolean list --json

hab helper timer create "Laundry" --duration 1:00:00
hab helper timer list --json

hab helper statistics create "Temp Average" --entity sensor.temperature --characteristic mean
hab helper statistics list --json
```

## Deletion and Verification

```bash
hab helper delete input_boolean.guest_mode
hab entity get input_boolean.guest_mode --json
```

If you are not sure which helper subtype command to use, start with `hab helper types --json`.

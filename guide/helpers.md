# Helper Commands

Use this topic for helper entities such as input booleans, counters, timers, groups, and template helpers.

## When to Use

- When creating Home Assistant helper entities.
- When choosing helper subtype commands for automation support.

## Prerequisites

- Check helper subtype capabilities first.
- Use JSON mode when helper output feeds automation logic.

## Discovery Sequence

```bash
hab helper types --json
hab helper input-boolean --help
hab helper statistics create --help
```

## Mutation Pattern

- Create helpers with explicit names and configuration flags.
- Verify generated entity IDs before using them in scripts or automations.

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

## Pitfalls

- Using the wrong helper subtype for the desired behavior.
- Assuming helper entity IDs from names without verifying actual output.

If you are not sure which helper subtype command to use, start with `hab helper types --json`.

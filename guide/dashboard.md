# Dashboard Creation Guide

Use this guide when building or refactoring Home Assistant dashboards.

For machine-readable output while inspecting dashboard resources, add `--json` to list/get commands.

## When to Use

- When creating a new dashboard from scratch.
- When iterating on view, section, and card composition.

## Prerequisites

- Discover devices and entities before composing views.
- Inspect existing dashboard structure before updates.

## Discovery Sequence

1. `hab dashboard list --json`
2. `hab dashboard get <dashboard_id> --json`
3. `hab dashboard view list <dashboard_id> --json`
4. `hab dashboard card list <dashboard_id> <view_index> --json`

## Mutation Pattern

- Use `dashboard patch` for field-level edits with an actual diff and verification.
- Existing `view`, `section`, and `card update --data` commands replace the selected object.
- Use YAML input (`-f` or heredoc) when creating complete views.

## Patch → Diff → Apply → Verify

```bash
hab schema dashboard patch --compact --json
hab dashboard patch my-dashboard --target /views/0/cards/0 --data '{"name":"Kitchen"}' --plan --json
hab dashboard patch my-dashboard --target /views/0/cards/0 --data '{"name":"Kitchen"}' --if-match 'sha256:<hash-from-base_revision>' --json
```

Use the preview's `base_revision` unchanged for apply. The target is an exact
JSON Pointer: `/views/0/sections/1/cards/2` addresses a section card, while
`/views/0/cards/0` addresses a regular-view card. Empty target selects the root.
No names are guessed and no missing parents are created. Whole-config freshness
checks protect index selectors against intervening reorderings.

The target must be an existing object. Escape `/` as `~1` and `~` as `~0` in
pointer segments; array indexes must be canonical nonnegative integers (`0`,
`1`, etc., without leading zeros).

`--data`/`--file` deep-merge objects and replace explicitly supplied arrays.
`--set '/name="Kitchen"'` uses a JSON value; `--remove /icon` deletes an object
field. Null remains a value. The operation order is merge, sets, removals.
Unrelated fields, including nested fields and exact JSON numbers, are retained.
Removing an already-missing object field is a no-op; missing intermediate
parents are errors. To remove array entries, replace the array explicitly.

Apply rechecks freshness immediately before saving and verifies stored JSON
afterward on the same connection. `status: verified` confirms read-back;
`status: noop` performs no save. On failure inspect `error.details.result`:
`saved: true` means save acknowledged, `saved: null` means uncertain. No write is
automatically retried. `reloaded: not_applicable` is explicit for storage saves.

HA has no conditional dashboard-save API: an edit after the final check can
still be overwritten. Verification proves stored JSON at `observed_at`, not
rendering, valid entity references, or custom resource readiness. Preview checks
local structure using live config; it does not check write permission or server
acceptance. Saving requires administrator permission. YAML dashboards may reject
saves, and patching does not expand generated strategy layouts. `--timeout`
bounds connection and requests after credential resolution; a timeout or
cancellation after sending a save does not prove it was unapplied.

`--diff-limit` bounds returned changes and exposes `change_count` and
`diff_complete`. This limits entry count, not bytes: replacing a large array can
still produce a large diff entry. Known credential fields and URL credentials
are redacted; secrets embedded in unlabelled prose cannot be reliably detected.
A conflict requires reinspection; copying a new hash without reviewing the new
diff defeats the freshness check.

## Look Beyond Entities - Explore Devices

When creating a dashboard for a specific purpose (e.g., a room, a function like "security"), don't limit yourself to searching for entities by name. Use `hab device list` to explore the devices in the system. Devices contain rich information including:

- All entities associated with the device
- Manufacturer and model information
- Device area assignment
- Configuration and diagnostic entities

This helps you discover related entities you might otherwise miss and understand the full capabilities of each device.

## Task-Focused Dashboards

When creating a dashboard focused on a specific task that involves a few devices (e.g., "Home Office", "Coffee Station", "Media Center"), include a **Maintenance section** alongside the primary controls. This section should contain:

- Battery levels for wireless devices
- Signal strength indicators
- Firmware update status
- Device connectivity states
- Any diagnostic entities relevant to the devices

This approach keeps users informed about the health of the devices supporting their task without cluttering the main interface. When something stops working, the maintenance section provides immediate visibility into potential issues.

## Respect Entity Categories

Entities have categories that indicate their intended purpose:

- **No category (primary)**: Main controls and states meant for regular user interaction
- **Diagnostic**: Entities for maintenance and troubleshooting (e.g., signal strength, battery level, firmware version)
- **Config**: Configuration entities for device settings (e.g., sensitivity levels, LED brightness)

When building dashboards:
- Group primary entities together for the main user interface
- Place diagnostic entities in a separate "Maintenance" or "Diagnostics" section
- Config entities typically belong in a dedicated settings area, not the main dashboard

This separation keeps dashboards clean and prevents users from accidentally changing configuration settings.

## Tile Card Features for Enhanced Control

Tile cards support features that provide additional control directly on the card. Consider using tile card features for:

- **Primary controls**: Light brightness slider, cover position, fan speed
- **Frequently used actions**: Toggle switches, quick actions

Avoid adding features to:
- Diagnostic entities
- Configuration entities
- Entities where simple state display is sufficient

Tile card features make important controls more accessible and visually prominent.

## Specialized Cards for Specific Domains

### Climate Entities
Use the **thermostat card** for climate entities. It provides:
- Current and target temperature display
- HVAC mode selection
- Temperature adjustment controls
- A visual representation that users intuitively understand

### Camera and Image Entities
Use **picture-entity cards** for camera and image entities:
- Hide the state (the image itself is the state)
- Hide the name unless the image context is ambiguous (most cameras and images are self-explanatory when viewed)
- Let the visual content speak for itself

## Using Badges for Global Information

Badges are ideal for displaying global data points that apply to an entire dashboard view. Good candidates include:

- Area temperature and humidity
- Security system status
- Weather conditions
- Presence/occupancy indicators
- General alerts or warnings

If the information is more specific to a subset of the dashboard, consider adding it to a section header instead of a badge. Badges work best for truly dashboard-wide context.

## Choosing the Right Graph Card

### Statistics Graph (for sensor entities)
Use **statistics-graph** cards when displaying sensor data over time:
- Automatically calculates and displays statistics (mean, min, max)
- Optimized for numerical sensor data
- Better performance for long time ranges

### History Graph (for other entity types)
Use **history-graph** cards for:
- Climate entity history (showing temperature changes alongside HVAC states)
- Binary sensor timelines
- State-based entities where you want to see state changes over time
- Any non-sensor entity where historical data is valuable

The history graph shows actual state changes as they occurred, which is more appropriate for non-numerical entities.

## Creating a Complete View at Once

Instead of creating sections and cards one by one, you can create an entire view with all its contents in a single command using YAML and a heredoc:

```bash
hab dashboard view create my-dashboard <<'EOF'
title: Living Room
icon: mdi:sofa
path: living-room
sections:
  - type: grid
    title: Lights
    cards:
      - type: tile
        entity: light.living_room_ceiling
        features:
          - type: light-brightness
      - type: tile
        entity: light.floor_lamp
      - type: tile
        entity: light.reading_lamp
  - type: grid
    title: Climate
    cards:
      - type: thermostat
        entity: climate.living_room
      - type: tile
        entity: sensor.living_room_temperature
      - type: tile
        entity: sensor.living_room_humidity
  - type: grid
    title: Media
    cards:
      - type: tile
        entity: media_player.tv
        features:
          - type: media-player-volume
      - type: tile
        entity: media_player.speaker
  - type: grid
    title: Maintenance
    cards:
      - type: tile
        entity: sensor.motion_sensor_battery
      - type: tile
        entity: sensor.temperature_sensor_battery
EOF
```

This approach is useful when:
- Building a new dashboard from scratch
- Migrating an existing dashboard configuration
- Creating templated views that can be reused

You can also save the view configuration to a file and use `-f`:

```bash
hab dashboard view create my-dashboard -f living-room-view.yaml
```

## Verification Commands

```bash
hab dashboard get <dashboard_id> --json
hab dashboard view list <dashboard_id> --json
hab dashboard card list <dashboard_id> <view_index> --json
```

## Pitfalls

- Editing cards before confirming view and section indexes.
- Building dashboards from entity name guesses without device context.

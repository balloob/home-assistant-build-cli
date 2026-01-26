# Dashboard Creation Guide

This guide provides best practices for building effective Home Assistant dashboards.

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

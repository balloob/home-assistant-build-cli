# Changelog

## [Unreleased]

### Saved logins survive network changes

The key that protects the saved login no longer depends on a network adapter's MAC address, so VPNs, virtual adapters (Hyper-V, WSL, Docker), and Wi-Fi changes no longer invalidate it. Existing logins are converted the next time they are read.

When a saved login still cannot be read, commands and `hab auth status` now say so and print the command to log in again, including the instance URL. The URL is stored next to the encrypted tokens for this purpose. Older `hab` versions cannot read the new credentials file.

## [1.9.0] - 2026-11-22

### App management

Manage Supervisor apps with `hab app`. List installed apps or browse the app store, inspect details, install, configure, update, start, stop, restart, and uninstall apps. Configuration accepts JSON or YAML from inline data, files, or stdin.

Preview changes with `--plan` or `--dry-run`. Stop, restart, and uninstall require confirmation. Uninstall preserves the app configuration directory unless `--remove-config` is set. Use `--timeout` for long operations and `hab guide apps` for workflow guidance.

## [1.8.0] - 2026-10-26

### Short resource IDs

Use unique prefixes of at least eight characters wherever commands select long opaque Home Assistant IDs. Discovery lists preserve full IDs and include `short_id`; `hab id list <resource>` covers additional resources such as users. ID flags and typed JSON/YAML references accept prefixes too. Exact IDs take priority, and ambiguous prefixes report the matching resources before any write.

Resolution respects category, to-do, trace, and repair scopes. Calendar event prefixes require the same `--start` and `--end` window used for discovery. Human-readable entity IDs and slugs remain exact.

### Entity registry updates

`hab entity update` can change entity IDs, names, icons, areas, labels, aliases, scoped categories, device classes, and hidden or disabled status. Preview changes with `--plan`; saved changes are read back and verified while preserving `unique_id`. Empty values clear overrides, and explicit false values reverse hidden or disabled status. Integration-specific options are excluded.

Entity ID changes do not rewrite references in dashboards, templates, automations, or external clients. Enabling an entity may require a reload or restart.

### Device metadata updates

Update device names, areas, labels, and disabled status with `hab device update`. The command supports live previews, clearing overrides, no-op detection, and read-back verification.

### Area, floor, and label metadata

Area, floor, and label updates support live previews, clearing supported fields, no-op detection, and read-back verification. Areas also support aliases, labels, pictures, and temperature/humidity sensor assignments. Floors support aliases.

### Bug fixes

- Resolve automation entity IDs to their configuration IDs when retrieving execution traces.


## [1.7.2] - 2026-10-02

- Consolidate dashboard patch and schema guidance in the maintained docs and remove the development milestone handoff.
- Correct command contracts, add compact searchable discovery, and introduce revision-checked dashboard patches with actual diffs, no-op detection, and read-back verification.
- Preserve shared integration-test lifecycle state so suite cleanup stops its owned Home Assistant process.

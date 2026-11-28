# ESPHome Workflows

Use this topic for ESPHome configuration, validation, builds, and uploads.

## When to Use

- When onboarding a new ESPHome device config.
- When patching, validating, and deploying firmware updates.

## Prerequisites

- A reachable ESPHome Device Builder (the ESPHome app's web UI) and valid auth context.
- A selected configuration (argument, `--device`, or saved context).

## Discovery Sequence

1. `hab esphome list --json`
2. `hab esphome info <configuration> --json`
3. `hab esphome config-read <configuration>`

## Mutation Pattern

- Validate before build/upload.
- Use `config-patch` for targeted YAML changes.
- `config-patch` and `update` refuse to write when the file changed since they read it; read it again and reapply the patch.
- Keep one configuration context active per workflow.

## Common Commands

```bash
hab esphome list --json
hab esphome context use <configuration>
hab esphome validate <configuration>
hab esphome build <configuration>
hab esphome upload <configuration>
hab esphome troubleshoot <configuration>   # offline device: DNS, mDNS and ping checks
```

## Boards and Create

- `boards <platform>` lists the Device Builder's board catalog; pass a board ID to `create --board`.
- `create --type basic` generates the config from the board's template. `--ssid`/`--psk` are stored as the shared `wifi_ssid`/`wifi_password` secrets.

## Verification Commands

```bash
hab esphome info <configuration> --json
hab esphome validate <configuration>
```

## Pitfalls

- Uploading firmware before validation succeeds.
- Running commands without setting configuration context.

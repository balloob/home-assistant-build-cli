# ESPHome Workflows

Use this topic for ESPHome configuration, validation, builds, and uploads.

## When to Use

- When onboarding a new ESPHome device config.
- When patching, validating, and deploying firmware updates.

## Prerequisites

- Reachable ESPHome dashboard and valid auth context.
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

## Dashboard Versions

- Both the ESPHome Device Builder (the dashboard in the ESPHome app since 2026.11) and the classic ESPHome dashboard work. hab uses the Device Builder's `/ws` API when it is there and falls back to the classic endpoints otherwise.
- `troubleshoot` needs the Device Builder.
- `boards`, `create` and `import` do not work on the Device Builder yet.

## Verification Commands

```bash
hab esphome info <configuration> --json
hab esphome validate <configuration>
```

## Pitfalls

- Uploading firmware before validation succeeds.
- Running commands without setting configuration context.
- Expecting `boards`, `create` or `import` to work on the ESPHome Device Builder; they need the classic dashboard for now.

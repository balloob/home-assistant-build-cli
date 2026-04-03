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
- Keep one configuration context active per workflow.

## Common Commands

```bash
hab esphome list --json
hab esphome context use <configuration>
hab esphome validate <configuration>
hab esphome build <configuration>
hab esphome upload <configuration>
```

## Verification Commands

```bash
hab esphome info <configuration> --json
hab esphome validate <configuration>
```

## Pitfalls

- Uploading firmware before validation succeeds.
- Running commands without setting configuration context.

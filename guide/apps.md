# App management

Apps are containerized applications managed by Supervisor. They were previously called add-ons. Use an administrator account on a Home Assistant installation with Supervisor.

## Discover

```bash
hab app list --json
hab app list --available --search mqtt --json
hab app get core_mosquitto --available --json
```

Use exact slugs from the list. `--available` reads the app store, including apps that are not installed. Lists also support `--brief`, `--count`, and `--limit`.

## Install and configure

```bash
hab app install core_mosquitto --plan
hab app install core_mosquitto
hab app get core_mosquitto --json
hab app configure core_mosquitto --data '{"boot":"auto","watchdog":true}' --plan
hab app configure core_mosquitto --data '{"boot":"auto","watchdog":true}'
hab app start core_mosquitto
hab app get core_mosquitto --json
```

`app get` includes the current options and their schema. Treat its output as sensitive: options can contain credentials.

Configuration accepts JSON or YAML through `--data`, `--file`, or stdin. Supported fields are `options`, `boot`, `auto_update`, `watchdog`, `ingress_panel`, `network`, `audio_input`, and `audio_output`. Omitted fields stay unchanged. Supply the complete `options` or `network` object when changing it. A null value resets options, network, or audio settings to their defaults.

Supervisor validates app-specific options. Changes do not automatically restart the app. Use `hab app restart <slug> --force` when needed.

## Operate and remove

```bash
hab backup create
hab app update core_mosquitto --plan
hab app update core_mosquitto --timeout 20m
hab app restart core_mosquitto --force
hab app stop core_mosquitto --force
hab app uninstall core_mosquitto --plan
hab app uninstall core_mosquitto --force
hab app list --json
```

Stop, restart, and uninstall require confirmation. `--force` skips it. Uninstall preserves the app configuration directory by default. Add `--remove-config` to delete that directory too.

All mutations support `--plan` and its alias `--dry-run`. Plans perform local checks without connecting to Home Assistant. They do not check the app state or validate its options against the server schema. Configuration values are hidden in plans.

Mutations wait up to ten minutes by default. Use `--timeout` to change the limit. A timeout or interrupted connection does not cancel an operation on Supervisor. Check the app state before retrying. Create a backup before updates or removal.

Use `hab guide management` for integration and repair flows, Assist, statistics, app administration, host and network settings, mounts, credentials, archive transfers, core settings, and media.

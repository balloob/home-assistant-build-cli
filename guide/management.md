# Instance management

Use `--json` for structured results. Most operations require an administrator account. App, host, network interface, and mount commands also require Supervisor. OS commands require Home Assistant OS.

Commands accept exact IDs from discovery results. Settings use `--data` (JSON or YAML), `--file`, or `--file -` for stdin. Send the fields supported by your Home Assistant version. The server validates their types and values. Do not include WebSocket envelope fields `id` or `type` in WebSocket payloads.

Mutations support `--plan` and `--dry-run`. These previews validate local input only. They do not contact the server or predict a configuration diff. One exception: `integration delete --plan` makes read-only calls to report its `impact`. It lists the devices and entities linked to the entry and, when no other enabled entry of its domain remains (ignored and disabled entries do not count), the loaded integrations whose manifest lists the domain under `dependencies` (those break). `after_dependencies` only orders setup and is not reported. If the server cannot be read, the plan still succeeds with `impact.status` set to `not_determined`. Destructive commands require confirmation; use `--force` for unattended execution. Use `--timeout 10m` for long Supervisor operations. A timeout can leave a mutation running. Inspect its result before retrying.

## Integration and repair flows

```bash
hab integration types --json
hab integration flow start <domain> --json
hab integration flow get <flow_id> --json
hab integration flow step <flow_id> --file answers.json --json
hab integration flow abort <flow_id>
hab integration flow list --json
hab integration options start <entry_id> --json
hab integration options step <flow_id> --data '{}' --json
hab integration flow start <domain> --data '{"entry_id":"<entry_id>"}' --json
hab integration update <entry_id> --data '{"title":"New name"}'
hab integration delete <entry_id> --plan --json
hab integration delete <entry_id> --force
```

`hab integration get <entry_id>` includes `dependent_integrations`: loaded integrations that require the entry's domain.

Read `type`, `step_id`, `data_schema`, `errors`, and `menu_options` in each flow result. Submit fields for that step. Use `{}` to confirm an empty form. A flow can require several steps. For a menu, send `next_step_id`. For an external step, open the returned URL and complete browser authentication. Then read the flow again. Stop at `create_entry` or `abort`. Flow errors can appear inside a successful HTTP response.

Reauthentication starts in Home Assistant when credentials fail. Find its pending flow with `integration flow list`, then resume it with `get` and `step`. Supply `entry_id` to `flow start` for reconfiguration. Options use their own `start`, `get`, `step`, and `abort` commands.

```bash
hab integration subentry list <entry_id> --json
hab integration subentry flow start <entry_id> <subentry_type> --json
hab integration subentry flow step <flow_id> --file answers.json --json
hab integration subentry flow start <entry_id> <subentry_type> --data '{"subentry_id":"<subentry_id>"}'
hab integration subentry update <entry_id> <subentry_id> --data '{"title":"New name"}'
hab integration subentry delete <entry_id> <subentry_id> --force
hab repairs list --json
hab repairs get <domain> <issue_id> --json
hab repairs flow start <domain> <issue_id> --json
hab repairs flow step <flow_id> --data '{}' --json
hab diagnostics download <entry_id> --output diagnostics.json
hab diagnostics download <entry_id> --device <device_id> --output device.json
```

Subentry and repair flows also provide `get` and `abort`. Diagnostics downloads contain the actual diagnostic report. Each integration controls which reports it supports.

## Assist and entity exposure

```bash
hab assist pipeline list --json
hab assist pipeline get <pipeline_id> --json
hab assist pipeline create --file pipeline.json
hab assist pipeline update <pipeline_id> --file pipeline.json
hab assist pipeline set-preferred <pipeline_id>
hab assist pipeline delete <pipeline_id> --force
hab assist languages --json
hab assist devices --json
hab assist runs list <pipeline_id> --json
hab assist runs get <pipeline_id> <pipeline_run_id> --json
hab assist exposure list --json
hab assist exposure set --data '{"assistants":["conversation"],"entity_ids":["light.kitchen"],"should_expose":true}'
hab assist exposure defaults get conversation --json
hab assist exposure defaults set conversation --data '{"expose_new":false}'
```

Pipeline creation requires `name`, `language`, `conversation_engine`, `conversation_language`, `stt_engine`, `stt_language`, `tts_engine`, `tts_language`, `tts_voice`, `wake_word_entity`, and `wake_word_id`. Set unused speech and wake word fields to `null`. For example, use `homeassistant` as the conversation engine and `en` for both language fields. To reuse a pipeline result as input, remove its `id`. Updates require the same fields as creation. Debug run history is limited to runs retained by Home Assistant.

Exposure assistant IDs are `conversation`, `cloud.alexa`, and `cloud.google_assistant`. The corresponding service must be available.

## Statistics

```bash
hab statistics query --data '{"start_time":"2026-01-01T00:00:00Z","period":"day","statistic_ids":["sensor.energy"],"types":["sum","state"]}' --json
hab statistics summary sensor.energy --data '{"fixed_period":{"start_time":"2026-01-01T00:00:00Z","end_time":"2026-02-01T00:00:00Z"}}' --json
hab statistics metadata list --data '{"statistic_ids":["sensor.energy"]}' --json
hab statistics validate --json
hab statistics metadata update sensor.energy --data '{"unit_of_measurement":"kWh","unit_class":"energy"}'
hab statistics adjust-sum sensor.energy --data '{"start_time":"2026-01-01T00:00:00Z","adjustment":1.5,"adjustment_unit_of_measurement":"kWh"}' --force
hab statistics clear --data '{"statistic_ids":["sensor.retired"]}' --force
hab statistics refresh-issues
```

Query periods are `5minute`, `hour`, `day`, `week`, `month`, and `year`. Optional fields include `end_time`, `units`, and `types`. A summary accepts one of `fixed_period`, `calendar`, or `rolling_window`. Unit metadata and sum corrections require their unit field, including an explicit `null` for unitless statistics. Use `validate` to inspect issues after a correction. Clearing without a time range removes all statistics for the selected IDs. There is no undo.

## Apps, host, network, and storage

```bash
hab app logs <slug> --json
hab app stats <slug> --json
hab app validate <slug> --file options.json
hab app security <slug> --data '{"protected":true}' --force
hab app rebuild <slug> --timeout 10m --force
hab app repository list --json
hab app repository add https://github.com/example/apps
hab app repository remove <repository_slug> --force
hab app refresh
hab host info --json
hab host options --data '{"hostname":"homeassistant"}'
hab host os info --json
hab host os update --timeout 10m --force
hab host os sync --force
hab host disks list --json
hab host disks usage <disk> --json
hab host datadisk list --json
hab host datadisk move <device> --timeout 10m --force
hab host reboot --force
hab host shutdown --force
hab network interfaces list --json
hab network interfaces scan <interface> --json
hab network interfaces update <interface> --file network.json --force
hab mount list --json
hab mount create --file mount.json
hab mount update <name> --file mount.json
hab mount reload <name>
hab mount options --data '{"default_backup_mount":"nas"}'
hab mount delete <name> --force
```

App validation accepts the options object directly. Protected mode uses the `protected` boolean. OS update optionally accepts `version`. Host options include `hostname`. OS configuration sync imports configuration from external media.

Network input follows the Supervisor interface schema. For DHCP, use `{"ipv4":{"method":"auto"}}`. Static IPv4 uses `method: static`, `address` as a list of CIDR addresses, `gateway`, and `nameservers`. Wi-Fi settings belong in `wifi`; use the interface discovery result to identify the correct device. A network change, OS update, reboot, or data disk move can disconnect this CLI.

Mounts require `name`, `type`, and `usage`. For NFS, add `server` and `path`; for CIFS, add `server`, `share`, and any required `username` and `password`. Usage is `backup`, `media`, or `share`. Updates require the full mount settings except `name`, which comes from the argument. Omit `name` from the update payload. Store passwords in a private input file.

## Users and application credentials

```bash
hab user list --json
hab user create --data '{"name":"Guest","group_ids":["system-users"],"local_only":true}' --json
hab user credentials create <user_id> --file credentials.json
hab user update <user_id> --data '{"is_active":false}'
hab user password <user_id> --file password.json --force
hab user username <user_id> --data '{"username":"guest"}'
hab user delete <user_id> --force
hab application-credential config --json
hab application-credential list --json
hab application-credential entry <config_entry_id> --json
hab application-credential create --file oauth.json
hab application-credential delete <application_credentials_id> --force
```

Creating a user does not create login credentials. The credentials file contains `username` and `password`. The password file for `user password` contains `{"password":"..."}`. Application credential input contains `domain`, `client_id`, `client_secret`, and optional `name`. Application credential results can contain client secrets. Keep these files and command output private.

## Backup archives, instance settings, and media

```bash
hab backup agents --json
hab backup download <backup_id> --agent <agent_id> --output backup.tar
hab backup upload backup.tar --agent <agent_id> --timeout 10m
hab system config get --json
hab system config update --data '{"location_name":"Home","time_zone":"Europe/Amsterdam","unit_system":"metric"}'
hab system config detect --json
hab media browse --json
hab media search music --data '{"media_content_id":"<source_id>"}' --json
hab media resolve <media_content_id> --json
hab media upload song.mp3 --media-content-id <directory_id>
hab media delete <media_content_id> --force
```

Backup download requires one agent. Upload can repeat `--agent` to select several destinations. Downloads preserve encryption unless `--password-file` supplies the backup password as plain text. File transfers use bounded memory. Downloads create a private file only after the transfer completes, and refuse to overwrite an existing path.

Core settings also include `latitude`, `longitude`, `elevation`, `radius`, `external_url`, `internal_url`, `currency`, `country`, and `language`. Unit systems are `metric` and `us_customary`. `detect` returns proposed location defaults; it does not apply them.

Media commands use IDs returned by `browse` and `search`. Only local media supports upload and deletion. Resolved URLs can be temporary. Downloads and uploads return a JSON result; binary contents never go to stdout.

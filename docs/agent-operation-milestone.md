# Agent operation: first milestone

Baseline: release **1.7.1**. This change implements the first milestone: command
contract corrections and compact discovery, followed by a complete dashboard
patch → diff → apply → verify workflow. Other work-order priorities remain staged
below; they are not implied by the new dashboard capability.

## Command-contract audit

The executable inventory is `cmd/schema_inventory.go`, with path-specific
side-effect and transport corrections in `cmd/schema_audit.go`. Contract tests
walk every registered executable, including factory-generated commands. New
executables must be added to the reviewed inventory; unknown operations require
a side-effect review. Existing schemas are custom contract descriptions, not
complete JSON Schema validators.

The resulting runtime index contains 290 executable paths, including Cobra help.

| Finding in 1.7.1 | Correction |
| --- | --- |
| `dashboard save-config`, `scene activate`, `automation create-from-blueprint` reported reads | Write classification and side-effect-free static preview |
| ESPHome build/upload/config writes/context changes reported reads | Write or destructive classification; serial probing also counts as write because esptool may reset the MCU |
| Repairs ignore/unignore and Thread preferred-dataset selection reported reads | Write classification and static preview |
| Trace, blueprint, integration, notification, system and entity transports inferred from parent families | Correct the command-specific REST/WebSocket/local/HTTP/serial routes |
| Generic mutation schemas invented `changed`, `id`, and `result` fields | Audited null/scalar/array/object shapes; explicitly open or opaque API-dependent payloads |
| Every list advertised brief/count regardless of flags | Advertise only supported variants |
| ESPHome validation default and stream previews described incorrectly | Default NDJSON, structured flag variant; mutation previews remain JSON envelopes |
| Unknown trailing schema path silently returned a parent | Fail explicitly with a nonzero exit |
| Generic plans implied an observed change and echoed credentials | Mark change detection/validation limits; redact credentials and opaque inputs |

`side_effect` describes the most consequential supported operation, not every
flag branch. For example, `esphome create --list-presets` is read-only but create
is classified as a write. Transport lists describe possible command backends,
not a promise that every invocation uses every backend. `esphome` denotes its
HTTP/WebSocket dashboard client, including optional HA ingress discovery.
Credential refresh and tool caches are ancillary local effects of existing
clients, not claims of changed HA configuration.

Input flags and positional arguments retain their existing invocation contract.
Standard JSON/YAML object inputs now expose an open `payload_schema`. API-owned
payloads marked `open: true` or `type: any` deliberately do not guarantee
undocumented fields, fixed shapes across flags, or version-independent content.
Legacy interactive authentication/help can print human-readable text even when
JSON was requested; use explicit noninteractive token-login arguments for a
single result envelope. This milestone does not redesign those prompts.

## Delivered review stages

| Stage | Implementation | Verification |
| --- | --- | --- |
| 1 — contracts/discovery | `cmd/schema*.go`, preview annotations and redaction | Every-executable inventory, critical classifications, strict paths, variants, schema size and compatibility tests; discovery CLI integration checks |
| 2 — dashboard workflow | `internal/dashboardpatch/`, `client/dashboard_patch.go`, `cmd/dashboard_patch.go` | Preserved fields/numbers, explicit pointers, actual differences, conflicts, no-ops, permission/capability failures, cancellation, timeout-after-apply |
| 3 — transport and handoff | Context-aware WebSocket requests, guides, this matrix and release note | HA integration, late-response/cancellation tests, legacy numeric-decoding compatibility, race tests and `go vet` |

## Interface

```sh
hab schema --index --search 'dashboard patch' --limit 10 --json
hab schema dashboard patch --compact --json
hab dashboard patch my-dashboard --target /views/0/cards/0 \
  --data '{"name":"Kitchen","tap_action":{"action":"more-info"}}' --plan --json
# Reuse data.base_revision from that preview, with the same patch:
hab dashboard patch my-dashboard --target /views/0/cards/0 \
  --data '{"name":"Kitchen","tap_action":{"action":"more-info"}}' \
  --if-match 'sha256:<base_revision_hash>' --json
```

Index responses include total matches, offset, `complete` (whether the final page
was reached), and `next_offset`. Search matches all case-insensitive words in the
canonical path, leaf aliases, and summary. A command path can scope the index.
The compact schema returns only the selected node and immediate child paths.
All schemas carry `schema_version` and `cli_version`; compact schemas also have
a content-derived `schema_id` and `envelope_ref: hab.envelope.v1`.

`--target` is an exact JSON Pointer into an existing object. It supports views,
badges represented as objects, sections, cards, and nested stack cards without
guessing by title or entity name. Array indexes must be canonical nonnegative
integers; no implicit last-item selection or parent scaffolding is performed.
Escape `/` as `~1` and `~` as `~0`. Empty target means the dashboard root.

Object input is deep-merged; arrays are replaced only when explicitly supplied;
null is a value. `--set '/field="JSON string"'` changes a field and `--remove
/field` removes an object member. Missing members are an idempotent removal;
missing intermediate parents are errors. Edits run in this order: merge, all
sets, then all removals. Removing array members implicitly is unsupported;
replace that array explicitly instead. Existing update/save commands keep their
replacement semantics for compatibility.

The full config is hashed, including fields outside the selected object. Apply
requires this precondition and checks it twice before sending one save. It then
reads the whole stored document on the same WebSocket connection and compares
the expected revision. Values remain lossless JSON numbers on this new path.

| Outcome | Meaning |
| --- | --- |
| `planned` | Live read + local pointer/merge checks + actual before/after diff; no write permission or server validation |
| `noop` | Desired state was already observed; no save; verified at that read |
| `verified` | Desired stored JSON was read back; `saved: true` means acknowledged, `saved: null` means acknowledgement was lost |
| `saved` with an error | Save acknowledged but verification unavailable or mismatched |
| `failed` | No save or definite rejection; inspect structured error code |
| `uncertain` | Save may have reached HA; desired state could not be verified |

Failures return a nonzero exit and the normal error envelope with the result at
`error.details.result`, a safe HA `cause_code` when available, and
`retryable: false`. No write is retried automatically. Reinspect uncertain or
conflicting outcomes before deciding whether to apply a new patch. Automatic
rollback is intentionally absent for this single-resource operation: a blind
restore could destroy a concurrent edit. `reloaded: not_applicable` is explicit;
storage dashboard saves need no separate reload.

Differences have paths, before/after values, and presence booleans so null and
deletion are distinguishable. `--diff-limit` bounds entry count and reports
`change_count`/`diff_complete`. It is not a byte limit for a changed large array.
Known credential keys and URL credentials are redacted in new diff results and
plans; revisions and verification use the original, unredacted values. Arbitrary
secrets embedded in unlabelled prose cannot be reliably identified.

## Capability matrix and API evidence

| Capability / environment | Evidence / support | Remaining limit |
| --- | --- | --- |
| Local discovery and contract inspection | No HA connection required | Open API-owned payloads are not exhaustive validators |
| HA Core 2026.9.4 storage dashboards | Live empty-hass integration; `lovelace/config`, `lovelace/config/save` | Save requires administrator permission |
| HA Core 2026.3.4 dashboard API | Source contract inspected; same read/save fields | Not runtime-tested in this milestone |
| Other HA versions | No new minimum version imposed; use compatible Lovelace APIs | Not certified; server errors are surfaced, not hidden |
| YAML/strategy dashboards and custom cards | Existing config may be readable; exact existing objects can be selected | YAML saves may be rejected; no generated-strategy expansion, frontend validation, or resource preflight |
| Concurrent writers | Full-document precondition and pre-save recheck | **No atomic conditional save**: a write after the last check can still be overwritten |
| Deadline/cancellation | Connection and dashboard requests bounded; pending requests cleaned up | Credential resolution precedes the bounded connection; cancellation cannot undo an already-sent save |
| Verification | Exact stored JSON observed with timestamp | Does not prove rendering, entity existence, resource loading, durable disk flush, or future state |

Verified upstream contracts:
- [HA 2026.9.4 Lovelace WebSocket API](https://github.com/home-assistant/core/blob/2026.9.4/homeassistant/components/lovelace/websocket.py)
- [HA 2026.3.4 Lovelace WebSocket API](https://github.com/home-assistant/core/blob/2026.3.4/homeassistant/components/lovelace/websocket.py)

The save request accepts `config` and optional `url_path`, with no revision,
ETag, conditional-write field, or validate-only action. The implementation does
not invent one. Default `lovelace` omits `url_path`, retaining the CLI's existing
mapping and HA's named-default fallback.

## Compatibility and release notes

- Existing command syntax and the response envelope are preserved. Compact/index
  formats are opt-in; the original full schema remains available.
- Corrected schema metadata is identified as `hab.command.v1.1`; consumers should
  key cached schemas by identity. Previous inaccurate inferred field lists are
  replaced with open/opaque shapes rather than false guarantees.
- Generic previews retain their envelope and `would_change` boolean for older
  consumers; it is an assumption when `change_detection` is `not_performed`.
  Positional values and opaque payload flags are now redacted.
- Legacy WebSocket calls retain float64 decoding. Only the new context-aware
  dashboard read/modify/write path opts into exact JSON numbers.
- No firmware upload, scene activation, or production HA mutation is required
  to inspect command contracts or use compact discovery.

## Measurements and reproducible checks

```sh
go test ./...
go vet ./...
go test -race ./client ./cmd ./input ./internal/dashboardpatch ./internal/redact
./test/run_integration_test.sh core registry dashboard misc
bash test/measure_discovery.sh /path/to/hab-1.7.1
```

The dashboard test uses 202 cards and reports inspection/preview/apply response
bytes, requests, and ten no-op repeats. A preview costs one config request; a
changed apply costs four (read, recheck, save, verify); a no-op costs one. Each
invocation reuses one connection. Request counts exclude authentication frames.
Compared with the old inspect → update → get pattern, this uses two CLI calls
instead of three but spends an extra HA request on the pre-save concurrency
check. The main context saving comes from returning only differences.

Measured locally against the 1.7.1 baseline and HA 2026.9.4 (JSON output bytes;
variable timestamps and dashboard IDs can change a few bytes):

| Workflow output | Bytes | HA config requests |
| --- | ---: | ---: |
| Baseline complete schema | 7,732,925 | 0 |
| Default command index (first 25 entries) | 6,220 | 0 |
| Search for dashboard patch | 973 | 0 |
| Dashboard patch full schema | 26,534 | 0 |
| Dashboard patch compact schema | 12,955 | 0 |
| Inspect 202-card dashboard | 40,935 | 1 |
| Preview two changed fields | 2,010 | 1 |
| Apply and verify those fields | 2,009 | 4 |

Ten repeated desired-state applications completed as no-ops with zero saves.
This is a bounded regression sample, not a production reliability guarantee.

## CI alignment

`.github/workflows/test.yml` builds and runs `go test ./...` with Go 1.22 on
Linux and Windows. Its Linux integration job uses Python 3.14 and runs the full
`test/run_integration_test.sh`. `.github/workflows/release.yml` builds Linux,
macOS and Windows binaries for amd64 and arm64 with the same Go version.

Go 1.22.12 Linux build, unit tests and vet passed locally, as did all six release
cross-builds and Windows amd64 test-binary compilation. Native Windows execution
must still run in the Windows CI job. Additional Go 1.26 race checks passed.
The shared integration helper now preserves its owned HA PID across repeated
imports, so the orchestrator can clean up its test process reliably.

The full CI integration command passed with **362 reported passes and zero
failures** on HA 2026.9.4, Go 1.22.12 and managed Python 3.14.3. The existing
runner counts some unavailable capabilities as passes/skips (for example,
Marketplace is absent on that HA version); this is not evidence of support for
those capabilities. New dashboard and discovery assertions are strict.

The first broad HA run found an environment failure also reproduced with the
unmodified baseline: `/api/services` returned HTTP 500 because system Python
3.14 lacked development headers and HA could not build `pymicro-vad`. A managed
Python 3.14 runtime with headers resolved it for the CI-aligned integration run;
the CLI does not suppress this failure or weaken its checks.

## Remaining implementation stages, in work-order priority

| Priority | Stage | Acceptance gate |
| --- | --- | --- |
| P1 | Extend surgical edits to automation/script components | Resolve configuration IDs; unique selectors; retained-field, conflict and actual-diff tests |
| P1 | Dependency and impact inspection | Identifiable references with direction; dynamic/unsupported references explicitly unknown; filesystem scanning only by explicit input |
| P1 | Execution explanations | Bounded trace evidence, failed steps/observed values/errors/timestamps; missing evidence distinct from a cause |
| P2 | Focused reads and batch execution | Projection, feasible pagination, shared sessions, completeness/freshness, per-operation failure |
| P2 | Reconciliation, jobs and event observation | Idempotence, resumability, bounded deadlines, explicit partial completion/retries, compensating recovery for multi-resource operations |
| P3 | Broader administration | Source-verified integration/options/subentry APIs, Assist readiness and resource preflight with version probes |

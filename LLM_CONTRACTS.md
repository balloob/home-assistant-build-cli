# LLM Contracts

This document defines the canonical machine-facing contracts used by `hab`.

## Response Envelope

All non-streaming JSON commands should return a single envelope with these top-level fields:

- `success`
- `operation`
- `resource_type`
- `data`
- `message`
- `partial_result`
- `warnings`
- `fallbacks_applied`
- `missing_sections`
- `verification_commands`
- `next_suggested_commands`
- `error`
- `metadata`

## Partial Results

Commands that degrade, omit requested sections, or fall back to alternate data sources must set:

- `partial_result: true`
- `warnings`: human-readable warnings
- `fallbacks_applied`: machine-readable fallback notes
- `missing_sections`: omitted sections when known

## Schema Contracts

`hab schema <command> --json` exposes both invocation metadata and output contracts.

Prefer `hab schema --index --search '<words>' --json` for bounded discovery and
`hab schema <command> --compact --json` for targeted schemas. These require no HA
connection. Unknown paths, including trailing path components, fail explicitly.

Command metadata is versioned as `hab.command.v1.1`. All schemas include the CLI
version; compact schemas additionally carry a content-derived `schema_id` and
reference the shared `hab.envelope.v1` envelope instead of repeating it. The
envelope above is unchanged. Breaking future contract formats require a new
major schema identifier. Open/opaque payload descriptions are explicit limits,
not permission to assume undocumented fields.

`side_effect` conservatively describes the most consequential supported branch.
`transports` lists possible command backends. `preview` distinguishes static
plans, command-specific plans, and actual live diffs. A plan with
`change_detection: not_performed` has not established whether a change is needed.

Output contracts include:

- `success_envelope`
- `error_envelope`
- `partial_envelope`
- `variants`
- `stream_events` for streaming commands

NDJSON streaming variants contain individual records, not an enclosing array.
Errors can end the stream with a standard error envelope. ESPHome mutation
previews are ordinary JSON envelopes.

When adding executable commands, update the reviewed inventory in
`cmd/schema_inventory.go` and review their side effects and transports in
`cmd/schema_audit.go`. Contract tests check coverage of every executable path.
These contracts describe command behavior; they are not JSON Schema validators.

## Verified Dashboard Patches

`hab dashboard patch` returns a bounded diff, full-config revisions, resolved
target pointer, observed timestamp, request count, and outcome flags. Apply
requires `--if-match`. Failure results live at `error.details.result`, preserving
the standard envelope and nonzero exit status. No-op is verified without a save;
an acknowledged save with failed read-back is not verified. Unknown save
acknowledgement is represented by `saved: null`. No automatic write retry or
blind rollback occurs. See the [dashboard guide](guide/dashboard.md) for patch
semantics and verification limits.

| Status | Meaning |
| --- | --- |
| `planned` | Live config read and local diff; no save or server validation |
| `noop` | Desired state observed; verified without a save |
| `verified` | Desired stored JSON read back; save acknowledgement may still be unknown |
| `saved` | Save acknowledged, but verification failed or was unavailable; returned with an error |
| `failed` | No save attempted or save definitely rejected |
| `uncertain` | Save may have reached HA; desired state could not be verified |

Workflow errors include `retryable: false` in `error.details` and a HA `cause_code` when
available. Reinspect conflicting or uncertain outcomes before applying a new
patch. Differences include before/after presence flags to distinguish missing
fields from null values. Redaction affects displayed values, not the config
used for revisions, saving, or verification.

## Guide Recipes

`hab guide <topic> --json` may include executable workflow recipes with:

- `required_capabilities`
- `inputs`
- `steps`
- `branches`
- `verification_steps`
- `recovery_steps`

Recipes are intended to let LLMs execute repeatable workflows without scraping prose.

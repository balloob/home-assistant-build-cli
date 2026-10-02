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

## Verified Dashboard Patches

`hab dashboard patch` returns a bounded diff, full-config revisions, resolved
target pointer, observed timestamp, request count, and outcome flags. Apply
requires `--if-match`. Failure results live at `error.details.result`, preserving
the standard envelope and nonzero exit status. No-op is verified without a save;
an acknowledged save with failed read-back is not verified. Unknown save
acknowledgement is represented by `saved: null`. No automatic write retry or
blind rollback occurs. See [the milestone audit and capability matrix](docs/agent-operation-milestone.md).

## Guide Recipes

`hab guide <topic> --json` may include executable workflow recipes with:

- `required_capabilities`
- `inputs`
- `steps`
- `branches`
- `verification_steps`
- `recovery_steps`

Recipes are intended to let LLMs execute repeatable workflows without scraping prose.

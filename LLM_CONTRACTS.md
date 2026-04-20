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

Output contracts include:

- `success_envelope`
- `error_envelope`
- `partial_envelope`
- `variants`
- `stream_events` for streaming commands

## Guide Recipes

`hab guide <topic> --json` may include executable workflow recipes with:

- `required_capabilities`
- `inputs`
- `steps`
- `branches`
- `verification_steps`
- `recovery_steps`

Recipes are intended to let LLMs execute repeatable workflows without scraping prose.

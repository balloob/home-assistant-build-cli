# Input and Output Patterns

Use this topic to keep agent workflows deterministic when reading output and sending payloads.

## When to Use

- Before writing JSON/YAML payload commands.
- When building command pipelines that parse CLI output.

## Prerequisites

- Use `--json` for parsed output.
- Choose one input style per command (`-d`, `-f`, or heredoc).

## Discovery Sequence

1. `hab schema --index --search '<task words>' --limit 10 --json`
2. `hab schema <command path> --compact --json`
3. `hab action docs <domain.action> --json` for live action inputs.
4. `hab guide <topic>` for workflow-specific patterns.

The index is bounded (25 entries by default) and returns `total`, `complete`,
and `next_offset`; pass that offset with the same search to continue. Compact
schemas contain invocation and payload contracts with a shared `envelope_ref`,
`schema_version`, `cli_version`, and content-derived `schema_id`. The original
full schema remains available without `--compact`. Unknown command paths fail.

Payloads marked `open` or `type: any` are not exhaustive schemas. `preview` is
`none`, `static`, `command_plan`, or `live_diff`. A static/command plan is not a
server validation or observed before/after diff; `change_detection` explains
whether `would_change` is merely assumed. Use `dashboard patch` for the verified
dashboard edit workflow.

## Mutation Pattern

- Use `-d` for short inline payloads.
- Use `-f` or heredoc for larger nested data.
- Keep payload format valid JSON or valid YAML, not mixed.

## Common Commands

```bash
hab overview --json
hab action docs light.turn_on --json
hab automation create evening_lights -f automation.yaml
hab dashboard view create my-dashboard <<'EOF'
title: Kitchen
sections: []
EOF
```

## Verification Commands

```bash
hab automation list --json
hab dashboard get my-dashboard --json
```

## Pitfalls

- Parsing text-mode output in automation logic.
- Mixing YAML indentation rules with JSON syntax in one payload.

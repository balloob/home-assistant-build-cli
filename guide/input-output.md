# Input and Output Patterns

Use this topic to keep agent workflows deterministic when reading output and sending payloads.

## When to Use

- Before writing JSON/YAML payload commands.
- When building command pipelines that parse CLI output.

## Prerequisites

- Use `--json` for parsed output.
- Choose one input style per command (`-d`, `-f`, or heredoc).

## Discovery Sequence

1. `hab <command> --help`
2. `hab action docs <domain.action> --json`
3. `hab guide <topic>` for workflow-specific patterns.

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

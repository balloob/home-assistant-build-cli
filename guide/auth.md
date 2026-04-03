# Authentication Workflows

Use this topic to establish and verify authentication before running workflows.

## When to Use

- At the start of a new environment/session.
- When commands start failing with auth or permission errors.

## Prerequisites

- Home Assistant URL reachable from the CLI environment.
- OAuth browser flow or a long-lived access token.

## Discovery Sequence

1. `hab auth status --json`
2. `hab auth discover --json`
3. `hab auth login`

## Mutation Pattern

- Authenticate once and reuse the stored session.
- Use `hab auth refresh --json` for long-running workflows.

## Common Commands

```bash
hab auth status --json
hab auth login
hab auth refresh --json
hab auth logout
```

## Verification Commands

```bash
hab auth status --json
hab overview --json
```

## Pitfalls

- Running write commands before checking session state.
- Mixing credentials from one Home Assistant instance with another.

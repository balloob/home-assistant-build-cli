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

## Multiple Instances (Contexts)

- Each context holds the URL and credentials of one Home Assistant instance; existing credentials are the `default` context.
- When told which context to use, pass `--context <name>` (or set `HAB_CONTEXT=<name>`) on every command instead of switching the current context.
- `hab auth context list --json` lists contexts and their URLs; `hab auth status --json` reports the context in use.
- `hab auth login --context <name>` adds a context; `hab auth logout --context <name>` removes only that context.
- An unknown context fails with `AUTH_REQUIRED` and lists the available contexts; it never falls back to another instance.
- `--context`/`HAB_CONTEXT` fails with a conflicting-credentials error while `HAB_URL` with `HAB_TOKEN` or `HAB_REFRESH_TOKEN` is set; unset those variables to use a named context (`hab auth login` warns about this).

## Common Commands

```bash
hab auth status --json
hab auth login
hab auth refresh --json
hab auth logout
hab auth context list --json
hab auth context use <name>
hab --context <name> auth status --json
```

## Verification Commands

```bash
hab auth status --json
hab overview --json
```

## Pitfalls

- Running write commands before checking session state.
- Mixing credentials from one Home Assistant instance with another.
- Switching the current context with `hab auth context use` when one command should target another instance; prefer `--context`.

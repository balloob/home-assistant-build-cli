# Operations and Maintenance

Use this topic for system operations such as backups, health checks, diagnostics, repairs, networking, and integration maintenance.

## When to Use

- Before disruptive operations like restore, restart, or network reconfiguration.
- During incident response or maintenance windows.

## Prerequisites

- Authenticated admin session.
- Recent backup and rollback strategy.

## Discovery Sequence

1. `hab system health --json`
2. `hab backup list --json`
3. `hab repairs list --json`
4. `hab diagnostics list --json`

## Mutation Pattern

- Create/verify backup before risky changes.
- Apply one operation at a time and re-check health.
- Track ignored repairs and revisit them intentionally.

## Common Commands

```bash
hab system health --json
hab backup create "Pre-maintenance"
hab repairs list --json
hab integration reload <entry_id>
hab network get --json
hab notification list --json
hab thread list --json
```

## Verification Commands

```bash
hab system health --json
hab diagnostics list --json
```

## Pitfalls

- Running restart/restore without a recent backup.
- Ignoring repairs without understanding their impact.

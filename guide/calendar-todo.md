# Calendar and To-Do Workflows

Use this topic for event scheduling and task-list automation.

## When to Use

- When creating, listing, and deleting calendar events.
- When adding and tracking to-do items in Home Assistant.

## Prerequisites

- Know target calendar and to-do entity IDs.
- Use ISO date/time formats for event fields.

## Discovery Sequence

1. `hab calendar list <calendar_entity_id> --json`
2. `hab todo lists --json`
3. `hab todo items <todo_entity_id> --json`

## Mutation Pattern

- Create one event or task at a time and keep returned IDs.
- Use `complete`/`uncomplete` transitions before deleting tasks.

## Common Commands

```bash
hab calendar list calendar.personal --json
hab calendar create calendar.personal --summary "Team Meeting" --start "2026-04-15T10:00:00" --end "2026-04-15T11:00:00"
hab todo lists --json
hab todo add todo.shopping_list "Buy milk"
hab todo complete todo.shopping_list <item_uid>
```

## Verification Commands

```bash
hab calendar list calendar.personal --json
hab todo items todo.shopping_list --json
```

## Pitfalls

- Using a to-do entity ID for calendar commands.
- Mixing all-day and timestamp event formats unintentionally.

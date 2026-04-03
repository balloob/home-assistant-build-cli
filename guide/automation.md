# Automation Commands

Use this topic for actions, automations, scripts, scenes, templates, and categories.

## When to Use

- For service calls and automation/script/scene workflows.
- When defining category scopes or template rendering in automations.

## Prerequisites

- Resolve target entity IDs first.
- Inspect action docs before action calls.

## Discovery Sequence

```bash
hab action list --json
hab action list light --json
hab action docs light.turn_on --json
hab action data --json
```

## Mutation Pattern

- Validate a single action call before embedding it into automations or scripts.
- Keep category scope explicit when creating and assigning categories.

## Execute Actions

```bash
hab action call light.turn_on --entity light.kitchen --data '{"brightness": 180}'
hab action call climate.set_temperature --entity climate.living_room --data '{"temperature": 22}'
```

## Automations, Scripts, Scenes

```bash
hab automation list --json
hab automation create evening_lights -f automation.yaml
hab automation trigger evening_lights

hab script list --json
hab script run morning_routine

hab scene list --json
hab scene activate scene.movie_mode
```

## Categories

Category scope is required and must be one of: `automation`, `script`, `scene`, `helpers`.

```bash
hab category list --scope automation --json
hab category create "Security" --scope automation
hab category assign <category_id> automation.evening_lights --scope automation
```

## Templates

```bash
hab template render "{{ states('sensor.outdoor_temperature') }}"
hab template render -f template.j2
```

## Verification Commands

```bash
hab automation list --json
hab script list --json
hab scene list --json
```

## Pitfalls

- Calling actions with incomplete payloads.
- Mixing category scopes and entity domains incorrectly.

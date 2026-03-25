# Automation Commands

Use this topic for actions, automations, scripts, scenes, templates, and categories.

## Discover First

```bash
hab action list --json
hab action list light --json
hab action docs light.turn_on --json
hab action data --json
```

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

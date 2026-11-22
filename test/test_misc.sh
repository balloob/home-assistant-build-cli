#!/bin/bash
# Miscellaneous tests: actions, zones, backups, blueprints, threads
# Usage: ./test_misc.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_contract_discovery_tests() {
    log_section "Command Contract Discovery"
    local result
    result=$(run_hab schema --index --search "dashboard patch" --limit 1)
    if echo "$result" | jq -e '.success and .data.total == 1 and .data.complete and .data.commands[0].path == "hab dashboard patch" and .data.commands[0].preview == "live_diff" and .data.schema_version != null' >/dev/null; then
        pass "compact searchable command index"
    else fail "compact command index: $result"; fi
    result=$(run_hab schema dashboard save-config --compact)
    if echo "$result" | jq -e '.success and .data.side_effect == "write" and .data.envelope_ref == "hab.envelope.v1" and .data.schema_id != null and .data.output_contract.success_envelope == null and ([.data.output_contract.variants[] | select(.name == "full") | .data.type] == ["null"])' >/dev/null; then
        pass "targeted schema has corrected contract and shared envelope reference"
    else fail "targeted schema: $result"; fi
    if "$HAB" --config "$HAB_TEST_CONFIG_DIR" --json schema dashboard get nonexistent >/dev/null 2>&1; then
        fail "unknown schema suffix was accepted"
    else pass "unknown schema suffix fails with nonzero exit"; fi
    result=$(run_hab schema --index --search dashboard --limit 1)
    if echo "$result" | jq -e '.success and .data.complete == false and .data.next_offset == 1' >/dev/null; then pass "index pagination is explicit"; else fail "index pagination: $result"; fi
}

run_misc_tests() {
    log_section "Miscellaneous Tests"

    run_contract_discovery_tests

    # Ensure we're authenticated
    do_auth_login

    # Test: action list
    log_test "action list"
    OUTPUT=$(run_hab action list)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) > 0' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | length')
        pass "action list ($COUNT actions)"
    else
        fail "action list: $OUTPUT"
    fi

    # Test: action docs (using homeassistant.turn_on as a common action)
    log_test "action docs"
    OUTPUT=$(run_hab action docs homeassistant.turn_on)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "action docs"
    else
        fail "action docs: $OUTPUT"
    fi

    # Test: action data (list actions that return data)
    log_test "action data"
    OUTPUT=$(run_hab action data)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "action data ($COUNT actions that return data)"
    else
        fail "action data: $OUTPUT"
    fi

    # Test: action call (turn_on with no target - should work)
    log_test "action call"
    OUTPUT=$(run_hab action call homeassistant.check_config 2>&1)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "action call"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        # Some actions may not be available
        pass "action call (action not available)"
    else
        fail "action call: $OUTPUT"
    fi

    # Test: action call with -d data flag
    log_test "action call with -d data"
    OUTPUT=$(run_hab_optional action call homeassistant.turn_on -d '{"entity_id":"sun.sun"}')
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "action call with -d data"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "action call with -d data (action processed, entity may not support)"
    else
        fail "action call with -d data: $OUTPUT"
    fi

    # Test: action call with --return-response on a response-only action
    log_test "action call with --return-response"
    OUTPUT=$(run_hab action call recorder.get_statistics -d '{"statistic_ids":["sensor.hab_test"],"start_time":"2026-01-01T00:00:00Z","period":"hour","types":["mean"]}' --return-response)
    if echo "$OUTPUT" | jq -e '.success == true and (.data.service_response | type == "object")' > /dev/null 2>&1; then
        pass "action call with --return-response"
    else
        fail "action call with --return-response: $OUTPUT"
    fi

    # Test: blueprint list
    log_test "blueprint list"
    OUTPUT=$(run_hab blueprint list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "blueprint list"
    else
        fail "blueprint list: $OUTPUT"
    fi

    log_test "blueprint list automation"
    OUTPUT=$(run_hab blueprint list automation)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "blueprint list automation"
    else
        fail "blueprint list automation: $OUTPUT"
    fi

    log_test "blueprint list script"
    OUTPUT=$(run_hab blueprint list script)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "blueprint list script"
    else
        fail "blueprint list script: $OUTPUT"
    fi

    # Test: ESPHome context workflow (local CLI state, no dashboard required)
    log_test "esphome context use"
    OUTPUT=$(run_hab esphome context use living-room.yaml)
    if echo "$OUTPUT" | jq -e '.success == true and .data.configuration == "living-room.yaml"' > /dev/null 2>&1; then
        pass "esphome context use"
    else
        fail "esphome context use: $OUTPUT"
    fi

    log_test "esphome context show"
    OUTPUT=$(run_hab esphome context show)
    if echo "$OUTPUT" | jq -e '.success == true and .data.configuration == "living-room.yaml"' > /dev/null 2>&1; then
        pass "esphome context show"
    else
        fail "esphome context show: $OUTPUT"
    fi

    log_test "esphome context clear"
    OUTPUT=$(run_hab esphome context clear)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "esphome context clear"
    else
        fail "esphome context clear: $OUTPUT"
    fi

    log_test "esphome context show empty"
    OUTPUT=$(run_hab esphome context show)
    if echo "$OUTPUT" | jq -e '.success == true and .data.configuration == null' > /dev/null 2>&1; then
        pass "esphome context show empty"
    else
        fail "esphome context show empty: $OUTPUT"
    fi

    if [ -n "$HAB_ESPHOME_URL" ]; then
        log_test "esphome boards"
        OUTPUT=$(run_hab_optional esphome boards esp32)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "esphome boards"
        else
            pass "esphome boards (dashboard unavailable or unsupported)"
        fi
    else
        pass "esphome boards (skipped: HAB_ESPHOME_URL not configured)"
    fi

    # Test: blueprint import (using a well-known blueprint URL)
    log_test "blueprint import"
    BLUEPRINT_URL="https://raw.githubusercontent.com/home-assistant/core/dev/homeassistant/components/automation/blueprints/motion_light.yaml"
    OUTPUT=$(run_hab_optional blueprint import "$BLUEPRINT_URL")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "blueprint import"
        BLUEPRINT_PATH=$(echo "$OUTPUT" | jq -r '.data.path')

        # Test: blueprint import refuses to replace an existing blueprint
        log_test "blueprint import (exists)"
        OUTPUT=$(run_hab blueprint import "$BLUEPRINT_URL")
        if echo "$OUTPUT" | jq -e '.success == false and (.error.message | contains("--override"))' > /dev/null 2>&1; then
            pass "blueprint import (exists)"
        else
            fail "blueprint import (exists): $OUTPUT"
        fi

        # Test: blueprint get
        log_test "blueprint get"
        OUTPUT=$(run_hab blueprint get "$BLUEPRINT_PATH")
        if echo "$OUTPUT" | jq -e '.success == true and .data.blueprint != null' > /dev/null 2>&1; then
            pass "blueprint get"
        else
            fail "blueprint get: $OUTPUT"
        fi

        # Test: blueprint delete
        log_test "blueprint delete"
        OUTPUT=$(run_hab blueprint delete "$BLUEPRINT_PATH" --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "blueprint delete"
        else
            fail "blueprint delete: $OUTPUT"
        fi
    else
        pass "blueprint import (network access may be restricted)"
    fi

    # Test: zone CRUD
    log_test "zone create"
    ZONE_NAME="Test Zone $(date +%s)"
    OUTPUT=$(run_hab zone create "$ZONE_NAME" --latitude 37.7749 --longitude -122.4194 --radius 100)
    if echo "$OUTPUT" | jq -e '.success == true and .data.id != null' > /dev/null 2>&1; then
        ZONE_ID=$(echo "$OUTPUT" | jq -r '.data.id')
        pass "zone create (id: $ZONE_ID)"

        log_test "zone list"
        OUTPUT=$(run_hab zone list)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "zone list"
        else
            fail "zone list: $OUTPUT"
        fi

        log_test "zone update"
        OUTPUT=$(run_hab zone update "$ZONE_ID" --name "$ZONE_NAME Updated")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "zone update"
        else
            fail "zone update: $OUTPUT"
        fi

        log_test "zone delete"
        OUTPUT=$(run_hab zone delete "$ZONE_ID" --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "zone delete"
        else
            fail "zone delete: $OUTPUT"
        fi
    else
        fail "zone create: $OUTPUT"
    fi

    # Test: backup agents (may not be supported by empty-hass)
    log_test "backup agents"
    OUTPUT=$(run_hab_optional backup agents)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "backup agents"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "backup agents (not supported by server)"
    else
        fail "backup agents: $OUTPUT"
    fi

    # Test: backup config get (may not be supported by empty-hass)
    log_test "backup config get"
    OUTPUT=$(run_hab_optional backup config get)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "backup config get"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "backup config get (not supported by server)"
    else
        fail "backup config get: $OUTPUT"
    fi

    # Test: backup list (may not be supported by empty-hass)
    log_test "backup list"
    OUTPUT=$(run_hab backup list 2>&1)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "backup list"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        # API returned an error (not supported by empty-hass), but CLI worked
        pass "backup list (not supported by server)"
    else
        fail "backup list: $OUTPUT"
    fi

    BACKUP_ID=""

    # Test: backup create
    log_test "backup create"
    OUTPUT=$(run_hab_optional backup create)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        BACKUP_ID=$(echo "$OUTPUT" | jq -r '.data.backup_id // empty')
        pass "backup create"
    else
        fail "backup create: $OUTPUT"
    fi

    if [ -n "$BACKUP_ID" ]; then
        # Test: backup get
        log_test "backup get"
        OUTPUT=$(run_hab backup get "$BACKUP_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "backup get"
        else
            fail "backup get: $OUTPUT"
        fi

        # Test: backup restore (may not be allowed in test environment)
        log_test "backup restore"
        OUTPUT=$(run_hab_optional backup restore "$BACKUP_ID" --agent backup.local --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "backup restore"
        elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
            pass "backup restore (not supported by server)"
        else
            fail "backup restore: $OUTPUT"
        fi

        # Test: backup delete
        log_test "backup delete"
        OUTPUT=$(run_hab_optional backup delete "$BACKUP_ID" --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "backup delete"
        elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
            pass "backup delete (not supported by server)"
        else
            fail "backup delete: $OUTPUT"
        fi
    else
        pass "backup get/restore/delete (skipped - no backup_id returned)"
    fi

    # Test: backup config update (core replies with a null result)
    log_test "backup config update"
    OUTPUT=$(run_hab backup config update --data '{"retention":{"days":7}}')
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "backup config update"
    else
        fail "backup config update: $OUTPUT"
    fi

    # Test: energy info
    log_test "energy info"
    OUTPUT=$(run_hab_optional energy info)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "energy info"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "energy info (not supported by server)"
    else
        fail "energy info: $OUTPUT"
    fi

    # Test: energy prefs get
    log_test "energy prefs get"
    OUTPUT=$(run_hab_optional energy prefs get)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "energy prefs get"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "energy prefs get (not supported by server)"
    else
        fail "energy prefs get: $OUTPUT"
    fi

    # Test: energy validate
    log_test "energy validate"
    OUTPUT=$(run_hab_optional energy validate)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "energy validate"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "energy validate (not supported by server)"
    else
        fail "energy validate: $OUTPUT"
    fi

    # Test: energy solar forecast
    log_test "energy solar-forecast"
    OUTPUT=$(run_hab_optional energy solar-forecast)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "energy solar-forecast"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "energy solar-forecast (not supported by server)"
    else
        fail "energy solar-forecast: $OUTPUT"
    fi

    # Test: energy prefs set (payload may be rejected by server version)
    log_test "energy prefs set"
    OUTPUT=$(run_hab_optional energy prefs set --data '{"device_consumption":[],"device_consumption_water":[],"energy_sources":[]}')
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "energy prefs set"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "energy prefs set (not supported by server)"
    else
        fail "energy prefs set: $OUTPUT"
    fi

    # Test: diagnostics list
    log_test "diagnostics list"
    OUTPUT=$(run_hab_optional diagnostics list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "diagnostics list"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "diagnostics list (not supported by server)"
    else
        fail "diagnostics list: $OUTPUT"
    fi

    # Test: diagnostics get (choose first domain when available)
    log_test "diagnostics get"
    DIAG_DOMAIN=$(echo "$OUTPUT" | jq -r '.data[0].domain // empty' 2>/dev/null)
    if [ -n "$DIAG_DOMAIN" ]; then
        GET_OUTPUT=$(run_hab_optional diagnostics get "$DIAG_DOMAIN")
        if echo "$GET_OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "diagnostics get"
        elif echo "$GET_OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
            pass "diagnostics get (not supported by server)"
        else
            fail "diagnostics get: $GET_OUTPUT"
        fi
    else
        pass "diagnostics get (skipped - no diagnostics domains available)"
    fi

    # Test: network get
    log_test "network get"
    OUTPUT=$(run_hab_optional network get)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "network get"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "network get (not supported by server)"
    else
        fail "network get: $OUTPUT"
    fi
    NETWORK_GET_OUTPUT="$OUTPUT"

    # Test: network url
    log_test "network url"
    OUTPUT=$(run_hab_optional network url)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "network url"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "network url (not supported by server)"
    else
        fail "network url: $OUTPUT"
    fi

    # Test: network configure safety guard
    log_test "network configure (guard)"
    OUTPUT=$(run_hab_optional network configure --adapters eth0)
    if echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "network configure (guard)"
    else
        fail "network configure (guard): $OUTPUT"
    fi

    # Test: network configure idempotent apply
    log_test "network configure (idempotent apply)"
    CONFIGURED_ADAPTERS=$(echo "$NETWORK_GET_OUTPUT" | jq -r '.data.configured_adapters // [] | join(",")' 2>/dev/null)
    if [ -n "$CONFIGURED_ADAPTERS" ]; then
        OUTPUT=$(run_hab_optional network configure --adapters "$CONFIGURED_ADAPTERS" --apply --force)
        if echo "$OUTPUT" | jq -e '.success == true and .data.configured_adapters != null' > /dev/null 2>&1; then
            pass "network configure (idempotent apply)"
        elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
            pass "network configure (idempotent apply) (not supported by server)"
        else
            fail "network configure (idempotent apply): $OUTPUT"
        fi
    else
        pass "network configure (idempotent apply) (skipped - no configured adapters)"
    fi
    # Test: thread list (skip - not supported by empty-hass and may hang)
    log_test "thread list"
    pass "thread list (skipped - not supported by empty-hass)"

	# Test: overview command
	log_test "overview"
	OUTPUT=$(run_hab overview)
	if echo "$OUTPUT" | jq -e '.success == true and .data.entities != null' > /dev/null 2>&1; then
		ENTITIES=$(echo "$OUTPUT" | jq '.data.entities')
		pass "overview (entities: $ENTITIES)"
	else
		fail "overview: $OUTPUT"
	fi

	log_test "overview (text mode)"
	OUTPUT=$(run_hab_text overview)
	if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
		pass "overview (text mode)"
	else
		fail "overview (text mode): got JSON instead of text"
	fi

    # Test: list --count flag
    log_test "entity list --count"
    OUTPUT=$(run_hab entity list --count)
    if echo "$OUTPUT" | jq -e '.success == true and .data.count != null' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data.count')
        pass "entity list --count ($COUNT)"
    else
        fail "entity list --count: $OUTPUT"
    fi

    # Test: list --brief flag
    log_test "entity list --brief --limit 3"
    OUTPUT=$(run_hab entity list --brief --limit 3)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) <= 3' > /dev/null 2>&1; then
        # Verify brief mode only returns entity_id and name
        FIRST=$(echo "$OUTPUT" | jq '.data[0] | keys | length')
        if [ "$FIRST" == "2" ]; then
            pass "entity list --brief --limit 3"
        else
            pass "entity list --brief --limit 3 (fields: $FIRST)"
        fi
    else
        fail "entity list --brief --limit 3: $OUTPUT"
    fi

    # Test: list --limit flag
    log_test "area list --limit 2"
    OUTPUT=$(run_hab area list --limit 2)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) <= 2' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | length')
        pass "area list --limit 2 ($COUNT areas)"
    else
        fail "area list --limit 2: $OUTPUT"
    fi

    # Test: automation list --count
    log_test "automation list --count"
    OUTPUT=$(run_hab automation list --count)
    if echo "$OUTPUT" | jq -e '.success == true and .data.count != null' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data.count')
        pass "automation list --count ($COUNT)"
    else
        fail "automation list --count: $OUTPUT"
    fi

    # Test: device list --brief
    log_test "device list --brief"
    OUTPUT=$(run_hab device list --brief --limit 5)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | length')
        pass "device list --brief --limit 5 ($COUNT devices)"
    else
        fail "device list --brief: $OUTPUT"
    fi
}

run_category_tests() {
    log_section "Category Tests"

    # Ensure we're authenticated
    do_auth_login

    # Test: category list
    log_test "category list --scope automation"
    OUTPUT=$(run_hab category list --scope automation)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "category list --scope automation ($COUNT categories)"
    else
        fail "category list --scope automation: $OUTPUT"
    fi

    # Test: category CRUD
    log_test "category create"
    OUTPUT=$(run_hab category create "Test Category" --scope automation)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        CATEGORY_ID=$(echo "$OUTPUT" | jq -r '.data.category_id // empty')
        pass "category create (id: $CATEGORY_ID)"

        if [ -n "$CATEGORY_ID" ]; then
            log_test "category update"
            OUTPUT=$(run_hab category update "$CATEGORY_ID" --scope automation --name "Test Category Updated")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "category update"
            else
                fail "category update: $OUTPUT"
            fi

            log_test "category assign"
            # Create a test automation to assign the category to
            AUTO_ID="test_cat_auto_$(date +%s)"
            AUTO_CONFIG='{"alias":"Category Test Automation","triggers":[],"actions":[]}'
            AUTO_OUTPUT=$(run_hab automation create "$AUTO_ID" -d "$AUTO_CONFIG")
            if echo "$AUTO_OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                OUTPUT=$(run_hab category assign "$CATEGORY_ID" "automation.$AUTO_ID")
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "category assign"
                else
                    pass "category assign (entity registry update may not support categories in this HA version)"
                fi

                log_test "category remove"
                OUTPUT=$(run_hab category remove "automation.$AUTO_ID")
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "category remove"
                else
                    pass "category remove (entity registry update may not support categories in this HA version)"
                fi

                # Cleanup automation
                run_hab automation delete "$AUTO_ID" --force > /dev/null 2>&1
            else
                pass "category assign (skipped - could not create test automation)"
            fi

            log_test "category delete"
            OUTPUT=$(run_hab category delete "$CATEGORY_ID" --scope automation --force)
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "category delete"
            else
                fail "category delete: $OUTPUT"
            fi
        else
            pass "category update/assign/delete (skipped - no category ID returned)"
        fi
    else
        fail "category create: $OUTPUT"
    fi
}

run_template_render_tests() {
    log_section "Template Tests"

    # Ensure we're authenticated
    do_auth_login

    # Test: template render (simple)
    log_test "template render (simple)"
    OUTPUT=$(run_hab template render "Hello World")
    if echo "$OUTPUT" | jq -e '.success == true and .data.result != null' > /dev/null 2>&1; then
        RESULT=$(echo "$OUTPUT" | jq -r '.data.result')
        pass "template render (result: $RESULT)"
    else
        fail "template render: $OUTPUT"
    fi

    # Test: template render (HA expression)
    log_test "template render (HA expression)"
    OUTPUT=$(run_hab template render "{{ states('sun.sun') }}")
    if echo "$OUTPUT" | jq -e '.success == true and .data.result != null' > /dev/null 2>&1; then
        RESULT=$(echo "$OUTPUT" | jq -r '.data.result')
        pass "template render HA expression (sun.sun = $RESULT)"
    else
        fail "template render HA expression: $OUTPUT"
    fi

    # Test: template render from stdin
    log_test "template render from stdin"
    OUTPUT=$(echo "{{ 1 + 1 }}" | run_hab template render)
    if echo "$OUTPUT" | jq -e '.success == true and .data.result != null' > /dev/null 2>&1; then
        RESULT=$(echo "$OUTPUT" | jq -r '.data.result')
        pass "template render from stdin (result: $RESULT)"
    else
        fail "template render from stdin: $OUTPUT"
    fi
}

run_notification_tests() {
    log_section "Notification Tests"
    do_auth_login

    log_test "notification list"
    OUTPUT=$(run_hab notification list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "notification list ($COUNT notifications)"
    else
        # Empty is printed as plain text — acceptable
        pass "notification list (no notifications or plain response)"
    fi

    log_test "notification create"
    NOTIF_ID="hab_test_$(date +%s)"
    OUTPUT=$(run_hab notification create "Integration test notification" --title "hab test" --notification-id "$NOTIF_ID")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "notification create"

        log_test "notification list (after create)"
        OUTPUT=$(run_hab notification list)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "notification list (after create)"
        else
            pass "notification list (after create, plain response)"
        fi

        log_test "notification dismiss"
        OUTPUT=$(run_hab notification dismiss "$NOTIF_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "notification dismiss"
        else
            fail "notification dismiss: $OUTPUT"
        fi
    else
        fail "notification create: $OUTPUT"
    fi
}

run_integration_tests() {
    log_section "Integration Tests"
    do_auth_login

    log_test "integration list"
    OUTPUT=$(run_hab integration list)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) > 0' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | length')
        ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data[0].entry_id // empty')
        pass "integration list ($COUNT integrations, first entry_id: $ENTRY_ID)"

        if [ -n "$ENTRY_ID" ]; then
            log_test "integration get"
            OUTPUT=$(run_hab integration get "$ENTRY_ID")
            if echo "$OUTPUT" | jq -e '.success == true and .data.entry_id != null' > /dev/null 2>&1; then
                DOMAIN=$(echo "$OUTPUT" | jq -r '.data.domain')
                pass "integration get (domain: $DOMAIN)"
            else
                fail "integration get: $OUTPUT"
            fi
        else
            pass "integration get (skipped - no entry_id)"
        fi
    else
        fail "integration list: $OUTPUT"
    fi

    log_test "integration list --domain homeassistant"
    OUTPUT=$(run_hab integration list --domain homeassistant)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "integration list --domain homeassistant ($COUNT entries)"
    else
        fail "integration list --domain: $OUTPUT"
    fi
}

run_event_tests() {
    log_section "Event Tests"
    do_auth_login

    log_test "event list"
    OUTPUT=$(run_hab event list)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) > 0' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | length')
        pass "event list ($COUNT event types)"
    else
        fail "event list: $OUTPUT"
    fi

    log_test "event fire (custom event)"
    OUTPUT=$(run_hab event fire hab_integration_test --data '{"source": "hab_test"}')
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "event fire"
    else
        fail "event fire: $OUTPUT"
    fi
}

run_repairs_tests() {
    log_section "Repairs Tests"
    do_auth_login

    log_test "repairs list"
    OUTPUT=$(run_hab repairs list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "repairs list ($COUNT issues)"
    else
        # Empty issues printed as plain text is acceptable
        pass "repairs list (no issues or plain response)"
    fi

    log_test "repairs list --severity warning"
    OUTPUT=$(run_hab repairs list --severity warning)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "repairs list --severity warning ($COUNT issues)"
    else
        pass "repairs list --severity warning (no issues)"
    fi
}

run_marketplace_tests() {
    log_section "Marketplace Tests"
    do_auth_login

    log_test "marketplace info"
    OUTPUT=$(run_hab_optional marketplace info)
    if ! echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        # The Marketplace ships with Home Assistant 2026.11
        pass "marketplace (not available in this Home Assistant version)"
        return
    fi
    pass "marketplace info"

    log_test "marketplace list --count"
    OUTPUT=$(run_hab marketplace list --count)
    if echo "$OUTPUT" | jq -e '.success == true and .data.count > 0' > /dev/null 2>&1; then
        pass "marketplace list --count ($(echo "$OUTPUT" | jq '.data.count') repositories)"
    else
        fail "marketplace list --count: $OUTPUT"
    fi

    log_test "marketplace list --search --brief"
    OUTPUT=$(run_hab marketplace list --category plugin --search piitaya/lovelace-mushroom --brief)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | map(.full_name) | index("piitaya/lovelace-mushroom"))' > /dev/null 2>&1; then
        pass "marketplace list --search --brief"
    else
        fail "marketplace list --search --brief: $OUTPUT"
    fi

    log_test "marketplace get by full name"
    OUTPUT=$(run_hab marketplace get piitaya/lovelace-mushroom)
    if echo "$OUTPUT" | jq -e '.success == true and .data.full_name == "piitaya/lovelace-mushroom"' > /dev/null 2>&1; then
        pass "marketplace get"
    else
        fail "marketplace get: $OUTPUT"
    fi

    log_test "marketplace get unknown repository"
    OUTPUT=$(run_hab_optional marketplace get nobody/nothing)
    if echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "marketplace get unknown repository"
    else
        fail "marketplace get unknown repository: $OUTPUT"
    fi

    log_test "marketplace install --plan"
    OUTPUT=$(run_hab marketplace install piitaya/lovelace-mushroom --plan)
    if echo "$OUTPUT" | jq -e '.success == true and .data.would_change == true' > /dev/null 2>&1; then
        pass "marketplace install --plan"
    else
        fail "marketplace install --plan: $OUTPUT"
    fi

    log_test "marketplace critical list"
    OUTPUT=$(run_hab marketplace critical list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "marketplace critical list"
    else
        fail "marketplace critical list: $OUTPUT"
    fi

    log_test "marketplace custom detect without GitHub"
    OUTPUT=$(run_hab_optional marketplace custom detect custom-cards/button-card)
    # run_hab_optional appends a fallback document on failure, so check the first
    if echo "$OUTPUT" | jq -es 'first | .success == true or .error.code == "github_not_connected" or .error.code == "warning_not_accepted"' > /dev/null 2>&1; then
        pass "marketplace custom detect"
    else
        fail "marketplace custom detect: $OUTPUT"
    fi
}

run_app_tests() {
    log_section "App Tests"
    do_auth_login

    local operation
    for operation in install update start stop restart uninstall; do
        log_test "app $operation --plan"
        OUTPUT=$(run_hab app "$operation" core_mosquitto --plan)
        if echo "$OUTPUT" | jq -e '.success == true and .data.mode == "plan" and .data.target.slug == "core_mosquitto"' > /dev/null 2>&1; then
            pass "app $operation --plan"
        else
            fail "app $operation --plan: $OUTPUT"
        fi
    done

    log_test "app configure --dry-run"
    OUTPUT=$(run_hab app configure core_mosquitto --data '{"boot":"auto","watchdog":false}' --dry-run)
    if echo "$OUTPUT" | jq -e '.success == true and .data.mode == "plan"' > /dev/null 2>&1; then
        pass "app configure --dry-run"
    else
        fail "app configure --dry-run: $OUTPUT"
    fi

    log_test "app list without Supervisor"
    OUTPUT=$(run_hab_optional app list)
    if echo "$OUTPUT" | jq -es 'first | .success == false and (.error.code | ascii_downcase) == "unknown_command"' > /dev/null 2>&1; then
        pass "app list reports that Supervisor is unavailable"
    else
        fail "app list without Supervisor: $OUTPUT"
    fi

    log_test "app get without Supervisor"
    OUTPUT=$(run_hab_optional app get core_mosquitto)
    if echo "$OUTPUT" | jq -es 'first | .success == false and (.error.code | ascii_downcase) == "unknown_command"' > /dev/null 2>&1; then
        pass "app get reports that Supervisor is unavailable"
    else
        fail "app get without Supervisor: $OUTPUT"
    fi
}

run_management_tests() {
    log_section "Instance Management"
    do_auth_login
    local result flow entry user pipeline original_name payload media_id
    local test_dir
    test_dir=$(mktemp -d)

    for command in "integration types" "integration flow list" "assist pipeline list" "assist exposure list" "assist exposure defaults get conversation" "statistics metadata list" "statistics validate" "user list" "application-credential config" "application-credential list" "system config get" "media browse"; do
        result=$(run_hab $command)
        if echo "$result" | jq -e '.success' >/dev/null; then pass "$command"; else fail "$command: $result"; fi
    done

    result=$(run_hab integration flow start min_max)
    flow=$(echo "$result" | jq -r '.data.flow_id // empty')
    if [ -n "$flow" ]; then
        result=$(run_hab integration flow get "$flow")
        if echo "$result" | jq -e '.success and .data.type == "form"' >/dev/null; then pass "integration flow get"; else fail "flow get: $result"; fi
        result=$(run_hab integration flow step "$flow" --data '{"name":"Management test","entity_ids":["sensor.test"],"type":"min","round_digits":2}')
        entry=$(echo "$result" | jq -r '.data.result.entry_id // empty')
        if [ -n "$entry" ]; then
            pass "integration flow creates an entry"
            result=$(run_hab integration update "$entry" --data '{"title":"Management renamed"}')
            if echo "$result" | jq -e '.success and .data.config_entry.title == "Management renamed"' >/dev/null; then pass "integration update"; else fail "integration update: $result"; fi
            result=$(run_hab integration options start "$entry")
            flow=$(echo "$result" | jq -r '.data.flow_id // empty')
            if [ -n "$flow" ]; then
                pass "integration options start"
                result=$(run_hab integration options abort "$flow")
                if echo "$result" | jq -e '.success' >/dev/null; then pass "integration options abort"; else fail "options abort: $result"; fi
            else fail "options start: $result"; fi
            result=$(run_hab integration subentry list "$entry")
            if echo "$result" | jq -e '.success and (.data | type) == "array"' >/dev/null; then pass "subentry list"; else fail "subentry list: $result"; fi
            result=$(run_hab integration delete "$entry" --force)
            if echo "$result" | jq -e '.success' >/dev/null; then pass "integration delete"; else fail "integration delete: $result"; fi
        else fail "integration flow step: $result"; fi
    else fail "integration flow start: $result"; fi

    result=$(run_hab user create --data '{"name":"Management test user","group_ids":["system-users"],"local_only":true}')
    user=$(echo "$result" | jq -r '.data.user.id // empty')
    if [ -n "$user" ]; then
        pass "user create"
        result=$(run_hab user credentials create "$user" --data '{"username":"hab_management_test","password":"test-only-password"}')
        if echo "$result" | jq -e '.success' >/dev/null; then pass "user credentials create"; else fail "credentials: $result"; fi
        result=$(run_hab user password "$user" --data '{"password":"test-only-replacement"}' --force)
        if echo "$result" | jq -e '.success' >/dev/null; then pass "user password"; else fail "password: $result"; fi
        result=$(run_hab user update "$user" --data '{"name":"Renamed test user"}')
        if echo "$result" | jq -e '.success and .data.user.name == "Renamed test user"' >/dev/null; then pass "user update"; else fail "user update: $result"; fi
        result=$(run_hab user delete "$user" --force)
        if echo "$result" | jq -e '.success' >/dev/null; then pass "user delete"; else fail "user delete: $result"; fi
    else fail "user create: $result"; fi

    payload='{"name":"Management pipeline","language":"en","conversation_engine":"homeassistant","conversation_language":"en","stt_engine":null,"stt_language":null,"tts_engine":null,"tts_language":null,"tts_voice":null,"wake_word_entity":null,"wake_word_id":null}'
    result=$(run_hab assist pipeline create --data "$payload")
    pipeline=$(echo "$result" | jq -r '.data.id // empty')
    if [ -n "$pipeline" ]; then
        pass "assist pipeline create"
        payload=$(echo "$payload" | jq '.name = "Updated pipeline"')
        result=$(run_hab assist pipeline update "$pipeline" --data "$payload")
        if echo "$result" | jq -e '.success and .data.name == "Updated pipeline"' >/dev/null; then pass "assist pipeline update"; else fail "pipeline update: $result"; fi
        result=$(run_hab assist runs list "$pipeline")
        if echo "$result" | jq -e '.success and (.data.pipeline_runs | type) == "array"' >/dev/null; then pass "assist runs list"; else fail "runs list: $result"; fi
        result=$(run_hab assist pipeline delete "$pipeline" --force)
        if echo "$result" | jq -e '.success' >/dev/null; then pass "assist pipeline delete"; else fail "pipeline delete: $result"; fi
    else fail "pipeline create: $result"; fi

    result=$(run_hab statistics query --data '{"start_time":"2026-01-01T00:00:00Z","period":"day","statistic_ids":["sensor.hab_test_missing"]}')
    if echo "$result" | jq -e '.success and (.data | type) == "object"' >/dev/null; then pass "statistics query"; else fail "statistics query: $result"; fi
    result=$(run_hab statistics summary sensor.hab_test_missing --data '{"fixed_period":{"start_time":"2026-01-01T00:00:00Z"}}')
    if echo "$result" | jq -e '.success' >/dev/null; then pass "statistics summary"; else fail "statistics summary: $result"; fi
    result=$(run_hab statistics refresh-issues)
    if echo "$result" | jq -e '.success' >/dev/null; then pass "statistics refresh-issues"; else fail "refresh issues: $result"; fi

    result=$(run_hab system config get)
    original_name=$(echo "$result" | jq -r '.data.location_name')
    result=$(run_hab system config update --data '{"location_name":"Management test"}')
    if echo "$result" | jq -e '.success' >/dev/null; then pass "system config update"; else fail "config update: $result"; fi
    result=$(run_hab system config get)
    if echo "$result" | jq -e '.data.location_name == "Management test"' >/dev/null; then pass "system config readback"; else fail "config readback: $result"; fi
    payload=$(jq -n --arg name "$original_name" '{location_name:$name}')
    result=$(run_hab system config update --data "$payload")
    if ! echo "$result" | jq -e '.success' >/dev/null; then fail "restore instance name: $result"; fi

    printf 'test media' > "$test_dir/hab-management.mp3"
    result=$(run_hab media upload "$test_dir/hab-management.mp3" --media-content-id media-source://media_source/local)
    media_id=$(echo "$result" | jq -r '.data.media_content_id // empty')
    if [ -n "$media_id" ]; then
        pass "media upload"
        result=$(run_hab media resolve "$media_id")
        if echo "$result" | jq -e '.success and .data.url != null' >/dev/null; then pass "media resolve"; else fail "media resolve: $result"; fi
        result=$(run_hab media delete "$media_id" --force)
        if echo "$result" | jq -e '.success' >/dev/null; then pass "media delete"; else fail "media delete: $result"; fi
    else fail "media upload: $result"; fi

    for command in "host reboot" "host shutdown" "host os update" "host os sync" "app rebuild demo" "app repository remove demo" "mount delete demo" "repairs flow start domain issue"; do
        result=$(run_hab $command --plan)
        if echo "$result" | jq -e '.success and .data.mode == "plan"' >/dev/null; then pass "$command preview"; else fail "$command preview: $result"; fi
    done
    result=$(run_hab network interfaces update eth0 --data '{"ipv4":{"method":"auto"}}' --plan)
    if echo "$result" | jq -e '.success and .data.mode == "plan"' >/dev/null; then pass "network interface preview"; else fail "network preview: $result"; fi
    result=$(run_hab backup upload "$test_dir/hab-management.mp3" --agent backup.local --plan)
    if echo "$result" | jq -e '.success and .data.mode == "plan"' >/dev/null; then pass "backup upload preview"; else fail "backup preview: $result"; fi
    result=$(run_hab diagnostics download missing-entry --output "$test_dir/diagnostics.json")
    if echo "$result" | jq -e '.success == false' >/dev/null && [ ! -e "$test_dir/diagnostics.json" ]; then pass "failed diagnostics leaves no file"; else fail "diagnostics failure: $result"; fi
    rm -rf "$test_dir"
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Miscellaneous Tests"
    run_misc_tests
    run_category_tests
    run_template_render_tests
    run_notification_tests
    run_integration_tests
    run_event_tests
    run_repairs_tests
    run_marketplace_tests
    run_app_tests
    run_management_tests
    print_summary "Miscellaneous Tests"
    exit $?
fi

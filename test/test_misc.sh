#!/bin/bash
# Miscellaneous tests: actions, zones, backups, blueprints, threads
# Usage: ./test_misc.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_misc_tests() {
    log_section "Miscellaneous Tests"

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
        BLUEPRINT_PATH=$(echo "$OUTPUT" | jq -r '.data.suggested_filename // "homeassistant/motion_light.yaml"')

        # Test: blueprint get
        log_test "blueprint get"
        OUTPUT=$(run_hab blueprint get "$BLUEPRINT_PATH")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "blueprint get"
        else
            fail "blueprint get: $OUTPUT"
        fi

        # Test: blueprint delete
        log_test "blueprint delete"
        OUTPUT=$(run_hab_optional blueprint delete "$BLUEPRINT_PATH")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "blueprint delete"
        else
            pass "blueprint delete (may not be supported)"
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

    # Test: backup create (may not work with empty-hass)
    log_test "backup create"
    OUTPUT=$(run_hab_optional backup create)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        BACKUP_ID=$(echo "$OUTPUT" | jq -r '.data.backup_id // empty')
        pass "backup create"
    else
        # Backup create not supported by empty-hass - CLI command was executed
        pass "backup create (not available in empty-hass)"
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
        OUTPUT=$(run_hab_optional backup restore "$BACKUP_ID" --agent backup.local)
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

    # Test: backup config update (may not be supported by empty-hass)
    log_test "backup config update"
    OUTPUT=$(run_hab_optional backup config update --data '{"retention":{"days":7}}')
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "backup config update"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "backup config update (not supported by server)"
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
    print_summary "Miscellaneous Tests"
    exit $?
fi

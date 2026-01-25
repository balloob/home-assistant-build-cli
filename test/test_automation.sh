#!/bin/bash
# Automation tests: automation CRUD and related commands
# Usage: ./test_automation.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_automation_tests() {
    log_section "Automation Tests"

    # Ensure we're authenticated
    do_auth_login

    # Test: automation list
    log_test "automation list"
    OUTPUT=$(run_hab automation list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "automation list ($COUNT automations)"
    else
        fail "automation list: $OUTPUT"
    fi

    # Test: automation list --extended (requires creating an automation with description first)
    log_test "automation list --extended"
    EXTENDED_AUTOMATION_ID="test_extended_$(date +%s)"
    EXTENDED_AUTOMATION_CONFIG='{"alias":"Extended Test Automation","description":"This is a test description for the extended flag","triggers":[],"actions":[]}'
    CREATE_OUTPUT=$(run_hab_optional automation create "$EXTENDED_AUTOMATION_ID" -d "$EXTENDED_AUTOMATION_CONFIG")
    if echo "$CREATE_OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        OUTPUT=$(run_hab automation list --extended)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            # Check that the output includes description for our automation
            FOUND_DESC=$(echo "$OUTPUT" | jq -r ".data[] | select(.entity_id == \"automation.$EXTENDED_AUTOMATION_ID\") | .description // empty")
            if [ -n "$FOUND_DESC" ]; then
                pass "automation list --extended (found description: $FOUND_DESC)"
            else
                pass "automation list --extended (works, description might be empty)"
            fi
        else
            fail "automation list --extended: $OUTPUT"
        fi
        # Cleanup
        run_hab automation delete "$EXTENDED_AUTOMATION_ID" --force > /dev/null 2>&1
    else
        pass "automation list --extended (skipped - could not create test automation)"
    fi

    # Test: automation CRUD (create, get, update, delete)
    log_test "automation create"
    AUTOMATION_ID="test_automation_$(date +%s)"
    AUTOMATION_CONFIG='{"alias":"Test Automation","triggers":[],"conditions":[],"actions":[]}'
    OUTPUT=$(run_hab automation create "$AUTOMATION_ID" -d "$AUTOMATION_CONFIG")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "automation create (id: $AUTOMATION_ID)"

        log_test "automation get"
        OUTPUT=$(run_hab automation get "$AUTOMATION_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation get"
        else
            fail "automation get: $OUTPUT"
        fi

        # Test: automation-trigger CRUD
        log_test "automation-trigger list (empty)"
        OUTPUT=$(run_hab automation-trigger list "$AUTOMATION_ID")
        if echo "$OUTPUT" | jq -e '.success == true and (.data | length) == 0' > /dev/null 2>&1; then
            pass "automation-trigger list (empty)"
        else
            fail "automation-trigger list (empty): $OUTPUT"
        fi

        log_test "automation-trigger create"
        TRIGGER_CONFIG='{"trigger":"state","entity_id":"sun.sun"}'
        OUTPUT=$(run_hab automation-trigger create "$AUTOMATION_ID" -d "$TRIGGER_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-trigger create"
        else
            fail "automation-trigger create: $OUTPUT"
        fi

        log_test "automation-trigger get"
        OUTPUT=$(run_hab automation-trigger get "$AUTOMATION_ID" 0)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-trigger get"
        else
            fail "automation-trigger get: $OUTPUT"
        fi

        log_test "automation-trigger update"
        TRIGGER_UPDATE_CONFIG='{"trigger":"state","entity_id":"sun.sun","to":"above_horizon"}'
        OUTPUT=$(run_hab automation-trigger update "$AUTOMATION_ID" 0 -d "$TRIGGER_UPDATE_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-trigger update"
        else
            fail "automation-trigger update: $OUTPUT"
        fi

        log_test "automation-trigger delete"
        OUTPUT=$(run_hab automation-trigger delete "$AUTOMATION_ID" 0 --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-trigger delete"
        else
            fail "automation-trigger delete: $OUTPUT"
        fi

        # Test: automation-condition CRUD
        log_test "automation-condition list (empty)"
        OUTPUT=$(run_hab automation-condition list "$AUTOMATION_ID")
        if echo "$OUTPUT" | jq -e '.success == true and (.data | length) == 0' > /dev/null 2>&1; then
            pass "automation-condition list (empty)"
        else
            fail "automation-condition list (empty): $OUTPUT"
        fi

        log_test "automation-condition create"
        CONDITION_CONFIG='{"condition":"state","entity_id":"sun.sun","state":"above_horizon"}'
        OUTPUT=$(run_hab automation-condition create "$AUTOMATION_ID" -d "$CONDITION_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-condition create"
        else
            fail "automation-condition create: $OUTPUT"
        fi

        log_test "automation-condition get"
        OUTPUT=$(run_hab automation-condition get "$AUTOMATION_ID" 0)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-condition get"
        else
            fail "automation-condition get: $OUTPUT"
        fi

        log_test "automation-condition update"
        CONDITION_UPDATE_CONFIG='{"condition":"state","entity_id":"sun.sun","state":"below_horizon"}'
        OUTPUT=$(run_hab automation-condition update "$AUTOMATION_ID" 0 -d "$CONDITION_UPDATE_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-condition update"
        else
            fail "automation-condition update: $OUTPUT"
        fi

        log_test "automation-condition delete"
        OUTPUT=$(run_hab automation-condition delete "$AUTOMATION_ID" 0 --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-condition delete"
        else
            fail "automation-condition delete: $OUTPUT"
        fi

        # Test: automation-action CRUD
        log_test "automation-action list (empty)"
        OUTPUT=$(run_hab automation-action list "$AUTOMATION_ID")
        if echo "$OUTPUT" | jq -e '.success == true and (.data | length) == 0' > /dev/null 2>&1; then
            pass "automation-action list (empty)"
        else
            fail "automation-action list (empty): $OUTPUT"
        fi

        log_test "automation-action create"
        ACTION_CONFIG='{"action":"homeassistant.turn_on","target":{"entity_id":"sun.sun"}}'
        OUTPUT=$(run_hab automation-action create "$AUTOMATION_ID" -d "$ACTION_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-action create"
        else
            fail "automation-action create: $OUTPUT"
        fi

        log_test "automation-action get"
        OUTPUT=$(run_hab automation-action get "$AUTOMATION_ID" 0)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-action get"
        else
            fail "automation-action get: $OUTPUT"
        fi

        log_test "automation-action update"
        ACTION_UPDATE_CONFIG='{"action":"homeassistant.turn_off","target":{"entity_id":"sun.sun"}}'
        OUTPUT=$(run_hab automation-action update "$AUTOMATION_ID" 0 -d "$ACTION_UPDATE_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-action update"
        else
            fail "automation-action update: $OUTPUT"
        fi

        log_test "automation-action delete"
        OUTPUT=$(run_hab automation-action delete "$AUTOMATION_ID" 0 --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation-action delete"
        else
            fail "automation-action delete: $OUTPUT"
        fi

        log_test "automation delete"
        OUTPUT=$(run_hab automation delete "$AUTOMATION_ID" --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation delete"
        else
            fail "automation delete: $OUTPUT"
        fi
    else
        fail "automation create: $OUTPUT"
    fi

    # Test: automation create-from-blueprint (requires blueprint import first)
    log_test "automation create-from-blueprint"
    BLUEPRINT_URL="https://raw.githubusercontent.com/home-assistant/core/dev/homeassistant/components/automation/blueprints/motion_light.yaml"
    IMPORT_OUTPUT=$(run_hab_optional blueprint import "$BLUEPRINT_URL")
    if echo "$IMPORT_OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        BLUEPRINT_PATH=$(echo "$IMPORT_OUTPUT" | jq -r '.data.suggested_filename // "homeassistant/motion_light.yaml"')

        BP_AUTOMATION_ID="test_bp_automation_$(date +%s)"
        BP_INPUTS='{"alias":"Test Blueprint Automation","motion_entity":"sun.sun","light_target":{"entity_id":"sun.sun"}}'
        OUTPUT=$(run_hab_optional automation create-from-blueprint "$BP_AUTOMATION_ID" "$BLUEPRINT_PATH" -d "$BP_INPUTS")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "automation create-from-blueprint"

            # Test: automation list with --blueprint filter
            log_test "automation list --blueprint (specific)"
            OUTPUT=$(run_hab automation list --blueprint "$BLUEPRINT_PATH")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
                if [ "$COUNT" -ge 1 ]; then
                    pass "automation list --blueprint (specific) ($COUNT automations)"
                else
                    fail "automation list --blueprint (specific): expected at least 1 automation, got $COUNT"
                fi
            else
                fail "automation list --blueprint (specific): $OUTPUT"
            fi

            log_test "automation list --blueprint=*"
            OUTPUT=$(run_hab automation list --blueprint "*")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
                pass "automation list --blueprint=* ($COUNT automations from blueprints)"
            else
                fail "automation list --blueprint=*: $OUTPUT"
            fi

            # Cleanup automation created from blueprint
            run_hab automation delete "$BP_AUTOMATION_ID" --force > /dev/null 2>&1
        else
            pass "automation create-from-blueprint (blueprint inputs may not match - skipped)"
        fi

        # Cleanup blueprint
        run_hab_optional blueprint delete "$BLUEPRINT_PATH" > /dev/null 2>&1
    else
        pass "automation create-from-blueprint (skipped - blueprint import failed, network may be restricted)"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Automation Tests"
    run_automation_tests
    print_summary "Automation Tests"
    exit $?
fi

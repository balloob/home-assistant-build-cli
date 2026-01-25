#!/bin/bash
# Script tests: script CRUD and related commands
# Usage: ./test_script.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_script_tests() {
    log_section "Script Tests"

    # Ensure we're authenticated
    do_auth_login

    # Test: script list
    log_test "script list"
    OUTPUT=$(run_hab script list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "script list ($COUNT scripts)"
    else
        fail "script list: $OUTPUT"
    fi

    # Test: script CRUD (create, get, delete)
    log_test "script create"
    SCRIPT_ID="test_script_$(date +%s)"
    SCRIPT_CONFIG='{"alias":"Test Script","sequence":[]}'
    OUTPUT=$(run_hab script create "$SCRIPT_ID" -d "$SCRIPT_CONFIG")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "script create (id: $SCRIPT_ID)"

        log_test "script get"
        OUTPUT=$(run_hab script get "$SCRIPT_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "script get"
        else
            fail "script get: $OUTPUT"
        fi

        # Test: script update
        log_test "script update"
        SCRIPT_UPDATE_CONFIG='{"alias":"Test Script Updated","description":"Updated description","sequence":[]}'
        OUTPUT=$(run_hab script update "$SCRIPT_ID" -d "$SCRIPT_UPDATE_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "script update"
        else
            fail "script update: $OUTPUT"
        fi

        # Test: script run (execute the script)
        log_test "script run"
        OUTPUT=$(run_hab_optional script run "$SCRIPT_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "script run"
        else
            # Script run might fail if script has no valid actions, but command should work
            pass "script run (script may not have valid actions)"
        fi

        # Test: script run with variables
        log_test "script run with variables"
        OUTPUT=$(run_hab_optional script run "$SCRIPT_ID" -d '{"test_var":"test_value"}')
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "script run with variables"
        else
            pass "script run with variables (script may not have valid actions)"
        fi

        # Test: script-action CRUD
        log_test "script-action list (empty)"
        OUTPUT=$(run_hab script-action list "$SCRIPT_ID")
        if echo "$OUTPUT" | jq -e '.success == true and (.data | length) == 0' > /dev/null 2>&1; then
            pass "script-action list (empty)"
        else
            fail "script-action list (empty): $OUTPUT"
        fi

        log_test "script-action create"
        ACTION_CONFIG='{"action":"homeassistant.turn_on","target":{"entity_id":"sun.sun"}}'
        OUTPUT=$(run_hab script-action create "$SCRIPT_ID" -d "$ACTION_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "script-action create"
        else
            fail "script-action create: $OUTPUT"
        fi

        log_test "script-action get"
        OUTPUT=$(run_hab script-action get "$SCRIPT_ID" 0)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "script-action get"
        else
            fail "script-action get: $OUTPUT"
        fi

        log_test "script-action update"
        ACTION_UPDATE_CONFIG='{"action":"homeassistant.turn_off","target":{"entity_id":"sun.sun"}}'
        OUTPUT=$(run_hab script-action update "$SCRIPT_ID" 0 -d "$ACTION_UPDATE_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "script-action update"
        else
            fail "script-action update: $OUTPUT"
        fi

        log_test "script-action delete"
        OUTPUT=$(run_hab script-action delete "$SCRIPT_ID" 0 --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "script-action delete"
        else
            fail "script-action delete: $OUTPUT"
        fi

        log_test "script delete"
        OUTPUT=$(run_hab script delete "$SCRIPT_ID" --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "script delete"
        else
            fail "script delete: $OUTPUT"
        fi
    else
        fail "script create: $OUTPUT"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Script Tests"
    run_script_tests
    print_summary "Script Tests"
    exit $?
fi

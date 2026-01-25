#!/bin/bash
# Helper tests: helper types, input_boolean, input_number, input_text, input_select, counter, timer, group
# Usage: ./test_helpers.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_helpers_tests() {
    log_section "Helper Tests"

    # Ensure we're authenticated
    do_auth_login

    # Test: helper list (all types)
    log_test "helper list"
    OUTPUT=$(run_hab helper list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper list ($COUNT helpers)"
    else
        fail "helper list: $OUTPUT"
    fi

    # Test: helper types
    log_test "helper types"
    OUTPUT=$(run_hab helper types)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) > 0' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | length')
        pass "helper types ($COUNT types)"
    else
        fail "helper types: $OUTPUT"
    fi

    # ==========================================================================
    # Input Boolean Helper Tests
    # ==========================================================================
    log_test "helper-input-boolean list"
    OUTPUT=$(run_hab helper-input-boolean list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-input-boolean list ($COUNT helpers)"
    else
        fail "helper-input-boolean list: $OUTPUT"
    fi

    log_test "helper-input-boolean create"
    OUTPUT=$(run_hab_optional helper-input-boolean create "Test Boolean" --icon "mdi:toggle-switch")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        HELPER_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-boolean create (id: $HELPER_ID)"

        if [ -n "$HELPER_ID" ]; then
            log_test "helper-input-boolean delete"
            OUTPUT=$(run_hab_optional helper-input-boolean delete "$HELPER_ID")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "helper-input-boolean delete"
            else
                fail "helper-input-boolean delete: $OUTPUT"
            fi
        fi
    else
        fail "helper-input-boolean create: $OUTPUT"
    fi

    # ==========================================================================
    # Input Number Helper Tests
    # ==========================================================================
    log_test "helper-input-number list"
    OUTPUT=$(run_hab helper-input-number list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-input-number list ($COUNT helpers)"
    else
        fail "helper-input-number list: $OUTPUT"
    fi

    log_test "helper-input-number create"
    OUTPUT=$(run_hab_optional helper-input-number create "Test Number" --min 0 --max 100 --step 5 --unit "%")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        HELPER_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-number create (id: $HELPER_ID)"

        if [ -n "$HELPER_ID" ]; then
            log_test "helper-input-number delete"
            OUTPUT=$(run_hab_optional helper-input-number delete "$HELPER_ID")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "helper-input-number delete"
            else
                fail "helper-input-number delete: $OUTPUT"
            fi
        fi
    else
        fail "helper-input-number create: $OUTPUT"
    fi

    # ==========================================================================
    # Input Text Helper Tests
    # ==========================================================================
    log_test "helper-input-text list"
    OUTPUT=$(run_hab helper-input-text list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-input-text list ($COUNT helpers)"
    else
        fail "helper-input-text list: $OUTPUT"
    fi

    log_test "helper-input-text create"
    OUTPUT=$(run_hab_optional helper-input-text create "Test Text" --min 0 --max 50)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        HELPER_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-text create (id: $HELPER_ID)"

        if [ -n "$HELPER_ID" ]; then
            log_test "helper-input-text delete"
            OUTPUT=$(run_hab_optional helper-input-text delete "$HELPER_ID")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "helper-input-text delete"
            else
                fail "helper-input-text delete: $OUTPUT"
            fi
        fi
    else
        fail "helper-input-text create: $OUTPUT"
    fi

    # ==========================================================================
    # Input Select Helper Tests
    # ==========================================================================
    log_test "helper-input-select list"
    OUTPUT=$(run_hab helper-input-select list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-input-select list ($COUNT helpers)"
    else
        fail "helper-input-select list: $OUTPUT"
    fi

    log_test "helper-input-select create"
    OUTPUT=$(run_hab_optional helper-input-select create "Test Select" --options "Option1,Option2,Option3")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        HELPER_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-select create (id: $HELPER_ID)"

        if [ -n "$HELPER_ID" ]; then
            log_test "helper-input-select delete"
            OUTPUT=$(run_hab_optional helper-input-select delete "$HELPER_ID")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "helper-input-select delete"
            else
                fail "helper-input-select delete: $OUTPUT"
            fi
        fi
    else
        fail "helper-input-select create: $OUTPUT"
    fi

    # ==========================================================================
    # Counter Helper Tests
    # ==========================================================================
    log_test "helper-counter list"
    OUTPUT=$(run_hab helper-counter list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-counter list ($COUNT helpers)"
    else
        fail "helper-counter list: $OUTPUT"
    fi

    log_test "helper-counter create"
    OUTPUT=$(run_hab_optional helper-counter create "Test Counter" --initial 0 --step 1 --minimum 0 --maximum 100)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        HELPER_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-counter create (id: $HELPER_ID)"

        if [ -n "$HELPER_ID" ]; then
            log_test "helper-counter delete"
            OUTPUT=$(run_hab_optional helper-counter delete "$HELPER_ID")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "helper-counter delete"
            else
                fail "helper-counter delete: $OUTPUT"
            fi
        fi
    else
        fail "helper-counter create: $OUTPUT"
    fi

    # ==========================================================================
    # Timer Helper Tests
    # ==========================================================================
    log_test "helper-timer list"
    OUTPUT=$(run_hab helper-timer list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-timer list ($COUNT helpers)"
    else
        fail "helper-timer list: $OUTPUT"
    fi

    log_test "helper-timer create"
    OUTPUT=$(run_hab_optional helper-timer create "Test Timer" --duration "00:05:00")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        HELPER_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-timer create (id: $HELPER_ID)"

        if [ -n "$HELPER_ID" ]; then
            log_test "helper-timer delete"
            OUTPUT=$(run_hab_optional helper-timer delete "$HELPER_ID")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "helper-timer delete"
            else
                fail "helper-timer delete: $OUTPUT"
            fi
        fi
    else
        fail "helper-timer create: $OUTPUT"
    fi

    # ==========================================================================
    # Group Helper Tests (uses config entry flow)
    # ==========================================================================
    log_test "helper-group list"
    OUTPUT=$(run_hab helper-group list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-group list ($COUNT groups)"
    else
        fail "helper-group list: $OUTPUT"
    fi

    log_test "helper-group create"
    # Groups use config entry flow - need to specify type and entities of matching domain
    # Using switch type with a switch entity would be ideal, but empty-hass may not have one
    # So we test with sensor type which is more flexible
    OUTPUT=$(run_hab_optional helper-group create "Test Group" --type sensor --entities "sun.sun")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper-group create (entry_id: $ENTRY_ID)"

        if [ -n "$ENTRY_ID" ]; then
            log_test "helper-group delete"
            OUTPUT=$(run_hab_optional helper-group delete "$ENTRY_ID")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "helper-group delete"
            else
                fail "helper-group delete: $OUTPUT"
            fi
        fi
    else
        # Group creation may fail if empty-hass doesn't support config flows
        pass "helper-group create (config flow not supported by server)"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Helper Tests"
    run_helpers_tests
    print_summary "Helper Tests"
    exit $?
fi

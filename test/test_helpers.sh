#!/bin/bash
# Helper tests: helper types, input_boolean, input_number, input_text, input_select, input_button, input_datetime, counter, timer, schedule, group
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
    OUTPUT=$(run_hab helper-input-boolean create "Test Boolean" --icon "mdi:toggle-switch")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        INPUT_BOOLEAN_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-boolean create (id: $INPUT_BOOLEAN_ID)"

        log_test "helper-input-boolean delete"
        OUTPUT=$(run_hab helper-input-boolean delete "$INPUT_BOOLEAN_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-input-boolean delete"
        else
            fail "helper-input-boolean delete: $OUTPUT"
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
    OUTPUT=$(run_hab helper-input-number create "Test Number" --min 0 --max 100 --step 5 --unit "%")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        INPUT_NUMBER_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-number create (id: $INPUT_NUMBER_ID)"

        log_test "helper-input-number delete"
        OUTPUT=$(run_hab helper-input-number delete "$INPUT_NUMBER_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-input-number delete"
        else
            fail "helper-input-number delete: $OUTPUT"
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
    OUTPUT=$(run_hab helper-input-text create "Test Text" --min 0 --max 50)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        INPUT_TEXT_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-text create (id: $INPUT_TEXT_ID)"

        log_test "helper-input-text delete"
        OUTPUT=$(run_hab helper-input-text delete "$INPUT_TEXT_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-input-text delete"
        else
            fail "helper-input-text delete: $OUTPUT"
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
    OUTPUT=$(run_hab helper-input-select create "Test Select" --options "Option1,Option2,Option3")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        INPUT_SELECT_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-select create (id: $INPUT_SELECT_ID)"

        log_test "helper-input-select delete"
        OUTPUT=$(run_hab helper-input-select delete "$INPUT_SELECT_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-input-select delete"
        else
            fail "helper-input-select delete: $OUTPUT"
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
    OUTPUT=$(run_hab helper-counter create "Test Counter" --initial 0 --step 1 --minimum 0 --maximum 100)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNTER_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-counter create (id: $COUNTER_ID)"

        log_test "helper-counter delete"
        OUTPUT=$(run_hab helper-counter delete "$COUNTER_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-counter delete"
        else
            fail "helper-counter delete: $OUTPUT"
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
    OUTPUT=$(run_hab helper-timer create "Test Timer" --duration "00:05:00")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TIMER_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-timer create (id: $TIMER_ID)"

        log_test "helper-timer delete"
        OUTPUT=$(run_hab helper-timer delete "$TIMER_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-timer delete"
        else
            fail "helper-timer delete: $OUTPUT"
        fi
    else
        fail "helper-timer create: $OUTPUT"
    fi

    # ==========================================================================
    # Input Button Helper Tests
    # ==========================================================================
    log_test "helper-input-button list"
    OUTPUT=$(run_hab helper-input-button list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-input-button list ($COUNT helpers)"
    else
        fail "helper-input-button list: $OUTPUT"
    fi

    log_test "helper-input-button create"
    OUTPUT=$(run_hab helper-input-button create "Test Button" --icon "mdi:button-pointer")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        INPUT_BUTTON_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-button create (id: $INPUT_BUTTON_ID)"

        log_test "helper-input-button delete"
        OUTPUT=$(run_hab helper-input-button delete "$INPUT_BUTTON_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-input-button delete"
        else
            fail "helper-input-button delete: $OUTPUT"
        fi
    else
        fail "helper-input-button create: $OUTPUT"
    fi

    # ==========================================================================
    # Input Datetime Helper Tests
    # ==========================================================================
    log_test "helper-input-datetime list"
    OUTPUT=$(run_hab helper-input-datetime list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-input-datetime list ($COUNT helpers)"
    else
        fail "helper-input-datetime list: $OUTPUT"
    fi

    log_test "helper-input-datetime create (date only)"
    OUTPUT=$(run_hab helper-input-datetime create "Test Date" --has-date --icon "mdi:calendar")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        INPUT_DATETIME_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-datetime create date only (id: $INPUT_DATETIME_ID)"

        log_test "helper-input-datetime delete (date)"
        OUTPUT=$(run_hab helper-input-datetime delete "$INPUT_DATETIME_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-input-datetime delete (date)"
        else
            fail "helper-input-datetime delete (date): $OUTPUT"
        fi
    else
        fail "helper-input-datetime create (date only): $OUTPUT"
    fi

    log_test "helper-input-datetime create (time only)"
    OUTPUT=$(run_hab helper-input-datetime create "Test Time" --has-time --icon "mdi:clock")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        INPUT_DATETIME_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-datetime create time only (id: $INPUT_DATETIME_ID)"

        log_test "helper-input-datetime delete (time)"
        OUTPUT=$(run_hab helper-input-datetime delete "$INPUT_DATETIME_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-input-datetime delete (time)"
        else
            fail "helper-input-datetime delete (time): $OUTPUT"
        fi
    else
        fail "helper-input-datetime create (time only): $OUTPUT"
    fi

    log_test "helper-input-datetime create (date and time)"
    OUTPUT=$(run_hab helper-input-datetime create "Test DateTime" --has-date --has-time)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        INPUT_DATETIME_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-input-datetime create date+time (id: $INPUT_DATETIME_ID)"

        log_test "helper-input-datetime delete (datetime)"
        OUTPUT=$(run_hab helper-input-datetime delete "$INPUT_DATETIME_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-input-datetime delete (datetime)"
        else
            fail "helper-input-datetime delete (datetime): $OUTPUT"
        fi
    else
        fail "helper-input-datetime create (date and time): $OUTPUT"
    fi

    # ==========================================================================
    # Schedule Helper Tests
    # ==========================================================================
    log_test "helper-schedule list"
    OUTPUT=$(run_hab helper-schedule list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-schedule list ($COUNT helpers)"
    else
        fail "helper-schedule list: $OUTPUT"
    fi

    log_test "helper-schedule create"
    OUTPUT=$(run_hab helper-schedule create "Test Schedule" --icon "mdi:calendar-clock")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        SCHEDULE_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "helper-schedule create (id: $SCHEDULE_ID)"

        log_test "helper-schedule delete"
        OUTPUT=$(run_hab helper-schedule delete "$SCHEDULE_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-schedule delete"
        else
            fail "helper-schedule delete: $OUTPUT"
        fi
    else
        fail "helper-schedule create: $OUTPUT"
    fi

    # ==========================================================================
    # Group Helper Tests (uses config entry flow via REST API)
    # ==========================================================================
    log_test "helper-group list"
    OUTPUT=$(run_hab helper-group list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-group list ($COUNT groups)"
    else
        fail "helper-group list: $OUTPUT"
    fi

    # Create input_number helpers to use in the group test
    log_test "helper-group create (setup: create input_numbers)"
    OUTPUT1=$(run_hab helper-input-number create "Group Test Number 1" --min 0 --max 100)
    OUTPUT2=$(run_hab helper-input-number create "Group Test Number 2" --min 0 --max 100)
    if echo "$OUTPUT1" | jq -e '.success == true' > /dev/null 2>&1 && \
       echo "$OUTPUT2" | jq -e '.success == true' > /dev/null 2>&1; then
        GROUP_NUM1_ID=$(echo "$OUTPUT1" | jq -r '.data.id // empty')
        GROUP_NUM2_ID=$(echo "$OUTPUT2" | jq -r '.data.id // empty')
        pass "helper-group create setup (created input_number.$GROUP_NUM1_ID, input_number.$GROUP_NUM2_ID)"

        log_test "helper-group create"
        # Groups use config entry flow - entities must match the group type domain
        OUTPUT=$(run_hab helper-group create "Test Sensor Group" --type sensor --entities "input_number.$GROUP_NUM1_ID,input_number.$GROUP_NUM2_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            GROUP_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
            pass "helper-group create (entry_id: $GROUP_ENTRY_ID)"

            log_test "helper-group delete"
            OUTPUT=$(run_hab helper-group delete "$GROUP_ENTRY_ID")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "helper-group delete"
            else
                fail "helper-group delete: $OUTPUT"
            fi
        else
            fail "helper-group create: $OUTPUT"
        fi

        # Cleanup the input_numbers we created for the group test
        run_hab helper-input-number delete "$GROUP_NUM1_ID" > /dev/null 2>&1
        run_hab helper-input-number delete "$GROUP_NUM2_ID" > /dev/null 2>&1
    else
        fail "helper-group create setup: failed to create input_numbers for group test"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Helper Tests"
    run_helpers_tests
    print_summary "Helper Tests"
    exit $?
fi

#!/bin/bash
# Template entity tests: alarm_control_panel, binary_sensor, button, image, number, select, sensor, switch
# Usage: ./test_template.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_template_tests() {
    log_section "Template Entity Tests"

    # Ensure we're authenticated
    do_auth_login

    # ==========================================================================
    # Template List Test
    # ==========================================================================
    log_test "helper-template list"
    OUTPUT=$(run_hab helper-template list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper-template list ($COUNT templates)"
    else
        fail "helper-template list: $OUTPUT"
    fi

    # ==========================================================================
    # Template Alarm Control Panel Tests
    # ==========================================================================
    log_test "helper-template create (alarm_control_panel)"
    OUTPUT=$(run_hab helper-template create "Test Alarm Panel" --type alarm_control_panel --state "{{ 'disarmed' }}")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TEMPLATE_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper-template create alarm_control_panel (entry_id: $TEMPLATE_ENTRY_ID)"

        log_test "helper-template delete (alarm_control_panel)"
        OUTPUT=$(run_hab helper-template delete "$TEMPLATE_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-template delete alarm_control_panel"
        else
            fail "helper-template delete alarm_control_panel: $OUTPUT"
        fi
    else
        fail "helper-template create alarm_control_panel: $OUTPUT"
    fi

    # ==========================================================================
    # Template Binary Sensor Tests
    # ==========================================================================
    log_test "helper-template create (binary_sensor)"
    OUTPUT=$(run_hab helper-template create "Test Binary Sensor" --type binary_sensor --state "{{ true }}")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TEMPLATE_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper-template create binary_sensor (entry_id: $TEMPLATE_ENTRY_ID)"

        log_test "helper-template delete (binary_sensor)"
        OUTPUT=$(run_hab helper-template delete "$TEMPLATE_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-template delete binary_sensor"
        else
            fail "helper-template delete binary_sensor: $OUTPUT"
        fi
    else
        fail "helper-template create binary_sensor: $OUTPUT"
    fi

    # ==========================================================================
    # Template Button Tests
    # ==========================================================================
    log_test "helper-template create (button)"
    OUTPUT=$(run_hab helper-template create "Test Button" --type button --press "homeassistant.check_config")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TEMPLATE_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper-template create button (entry_id: $TEMPLATE_ENTRY_ID)"

        log_test "helper-template delete (button)"
        OUTPUT=$(run_hab helper-template delete "$TEMPLATE_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-template delete button"
        else
            fail "helper-template delete button: $OUTPUT"
        fi
    else
        fail "helper-template create button: $OUTPUT"
    fi

    # ==========================================================================
    # Template Image Tests
    # ==========================================================================
    log_test "helper-template create (image)"
    OUTPUT=$(run_hab helper-template create "Test Image" --type image --url "https://example.com/image.jpg")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TEMPLATE_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper-template create image (entry_id: $TEMPLATE_ENTRY_ID)"

        log_test "helper-template delete (image)"
        OUTPUT=$(run_hab helper-template delete "$TEMPLATE_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-template delete image"
        else
            fail "helper-template delete image: $OUTPUT"
        fi
    else
        fail "helper-template create image: $OUTPUT"
    fi

    # ==========================================================================
    # Template Number Tests
    # ==========================================================================
    log_test "helper-template create (number)"
    OUTPUT=$(run_hab helper-template create "Test Number" --type number --state "{{ 50 }}" --min 0 --max 100 --step 5 --set-value "input_number.set_value")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TEMPLATE_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper-template create number (entry_id: $TEMPLATE_ENTRY_ID)"

        log_test "helper-template delete (number)"
        OUTPUT=$(run_hab helper-template delete "$TEMPLATE_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-template delete number"
        else
            fail "helper-template delete number: $OUTPUT"
        fi
    else
        fail "helper-template create number: $OUTPUT"
    fi

    # ==========================================================================
    # Template Select Tests
    # ==========================================================================
    log_test "helper-template create (select)"
    OUTPUT=$(run_hab helper-template create "Test Select" --type select --state "{{ 'option1' }}" --options "option1,option2,option3" --select-option "input_select.select_option")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TEMPLATE_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper-template create select (entry_id: $TEMPLATE_ENTRY_ID)"

        log_test "helper-template delete (select)"
        OUTPUT=$(run_hab helper-template delete "$TEMPLATE_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-template delete select"
        else
            fail "helper-template delete select: $OUTPUT"
        fi
    else
        fail "helper-template create select: $OUTPUT"
    fi

    # ==========================================================================
    # Template Sensor Tests
    # ==========================================================================
    log_test "helper-template create (sensor)"
    OUTPUT=$(run_hab helper-template create "Test Sensor" --type sensor --state "{{ 42 }}" --unit "°C" --device-class temperature)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TEMPLATE_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper-template create sensor (entry_id: $TEMPLATE_ENTRY_ID)"

        log_test "helper-template delete (sensor)"
        OUTPUT=$(run_hab helper-template delete "$TEMPLATE_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-template delete sensor"
        else
            fail "helper-template delete sensor: $OUTPUT"
        fi
    else
        fail "helper-template create sensor: $OUTPUT"
    fi

    # ==========================================================================
    # Template Switch Tests
    # ==========================================================================
    log_test "helper-template create (switch)"
    OUTPUT=$(run_hab helper-template create "Test Switch" --type switch --state "{{ false }}")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TEMPLATE_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper-template create switch (entry_id: $TEMPLATE_ENTRY_ID)"

        log_test "helper-template delete (switch)"
        OUTPUT=$(run_hab helper-template delete "$TEMPLATE_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper-template delete switch"
        else
            fail "helper-template delete switch: $OUTPUT"
        fi
    else
        fail "helper-template create switch: $OUTPUT"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Template Entity Tests"
    run_template_tests
    print_summary "Template Entity Tests"
    exit $?
fi

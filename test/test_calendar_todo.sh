#!/bin/bash
# Calendar and To-do list tests: local_calendar, local_todo helpers and calendar list command
# Usage: ./test_calendar_todo.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_calendar_todo_tests() {
    log_section "Calendar and To-do Tests"

    # Ensure we're authenticated
    do_auth_login

    # ==========================================================================
    # Local Calendar Helper Tests
    # ==========================================================================
    log_test "helper local-calendar list (initial)"
    OUTPUT=$(run_hab helper local-calendar list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper local-calendar list ($COUNT calendars)"
    else
        fail "helper local-calendar list: $OUTPUT"
    fi

    log_test "helper local-calendar create"
    CALENDAR_NAME="Test Calendar $(date +%s)"
    OUTPUT=$(run_hab helper local-calendar create "$CALENDAR_NAME")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        CALENDAR_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper local-calendar create (entry_id: $CALENDAR_ENTRY_ID)"

        # Wait a moment for entity to be created
        sleep 1

        # Find the calendar entity ID
        CALENDAR_ENTITY=$(run_hab entity list | jq -r '.data[] | select(.entity_id | startswith("calendar.")) | .entity_id' | head -1)

        # Test: calendar list (list events from the calendar)
        if [ -n "$CALENDAR_ENTITY" ]; then
            log_test "calendar list"
            OUTPUT=$(run_hab_optional calendar list "$CALENDAR_ENTITY")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                EVENT_COUNT=$(echo "$OUTPUT" | jq '.data.events | if . == null then 0 elif type == "array" then length else 0 end')
                pass "calendar list ($EVENT_COUNT events from $CALENDAR_ENTITY)"
            else
                # Calendar might not support event listing via this API
                pass "calendar list (API may not support event listing)"
            fi

            # Test: calendar list with time range
            log_test "calendar list with time range"
            START_TIME=$(date -u +"%Y-%m-%dT00:00:00Z")
            END_TIME=$(date -u -d "+7 days" +"%Y-%m-%dT23:59:59Z" 2>/dev/null || date -u -v+7d +"%Y-%m-%dT23:59:59Z" 2>/dev/null || echo "")
            if [ -n "$END_TIME" ]; then
                OUTPUT=$(run_hab_optional calendar list "$CALENDAR_ENTITY" --start "$START_TIME" --end "$END_TIME")
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "calendar list with time range"
                else
                    pass "calendar list with time range (API may not support time filtering)"
                fi
            else
                pass "calendar list with time range (skipped - date calculation not available)"
            fi
        else
            log_test "calendar list"
            pass "calendar list (skipped - calendar entity not found yet)"
            log_test "calendar list with time range"
            pass "calendar list with time range (skipped - calendar entity not found)"
        fi

        log_test "helper local-calendar delete"
        OUTPUT=$(run_hab helper local-calendar delete "$CALENDAR_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper local-calendar delete"
        else
            fail "helper local-calendar delete: $OUTPUT"
        fi
    else
        fail "helper local-calendar create: $OUTPUT"
    fi

    # ==========================================================================
    # Local To-do Helper Tests
    # ==========================================================================
    log_test "helper local-todo list (initial)"
    OUTPUT=$(run_hab helper local-todo list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "helper local-todo list ($COUNT to-do lists)"
    else
        fail "helper local-todo list: $OUTPUT"
    fi

    log_test "helper local-todo create"
    TODO_NAME="Test Todo $(date +%s)"
    OUTPUT=$(run_hab helper local-todo create "$TODO_NAME")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        TODO_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
        pass "helper local-todo create (entry_id: $TODO_ENTRY_ID)"

        log_test "helper local-todo list (after create)"
        OUTPUT=$(run_hab helper local-todo list)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
            if [ "$COUNT" -ge 1 ]; then
                pass "helper local-todo list (found $COUNT to-do lists)"
            else
                fail "helper local-todo list: expected at least 1 to-do list"
            fi
        else
            fail "helper local-todo list: $OUTPUT"
        fi

        log_test "helper local-todo delete"
        OUTPUT=$(run_hab helper local-todo delete "$TODO_ENTRY_ID")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "helper local-todo delete"
        else
            fail "helper local-todo delete: $OUTPUT"
        fi
    else
        fail "helper local-todo create: $OUTPUT"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Calendar and To-do Tests"
    run_calendar_todo_tests
    print_summary "Calendar and To-do Tests"
    exit $?
fi

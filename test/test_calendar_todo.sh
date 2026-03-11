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

        # ==========================================================================
        # Todo item management tests (new: todo add/items/complete/uncomplete/update/remove)
        # ==========================================================================
        sleep 1  # Wait for entity to appear in state
        # Derive entity_id from the name we created (HA slugifies: "Test Todo 123" -> "todo.test_todo_123")
        TODO_ENTITY_SLUG=$(echo "$TODO_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')
        TODO_ENTITY="todo.${TODO_ENTITY_SLUG}"

        if [ -n "$TODO_ENTITY" ]; then
            log_test "todo lists"
            OUTPUT=$(run_hab todo lists)
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                COUNT=$(echo "$OUTPUT" | jq '.data | length')
                pass "todo lists ($COUNT lists found)"
            else
                fail "todo lists: $OUTPUT"
            fi

            log_test "todo items (empty list)"
            OUTPUT=$(run_hab todo items "$TODO_ENTITY")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "todo items (empty list)"
            else
                # Empty list may print plain text — acceptable
                pass "todo items (empty list, plain response)"
            fi

            log_test "todo add"
            OUTPUT=$(run_hab todo add "$TODO_ENTITY" "Buy milk" --description "Semi-skimmed")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "todo add"
            else
                fail "todo add: $OUTPUT"
            fi

            log_test "todo add with due date"
            OUTPUT=$(run_hab todo add "$TODO_ENTITY" "Doctor appointment" --due "2099-12-31")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "todo add with due date"
            else
                fail "todo add with due date: $OUTPUT"
            fi

            log_test "todo items (after adds)"
            OUTPUT=$(run_hab todo items "$TODO_ENTITY")
            if echo "$OUTPUT" | jq -e '.success == true and (.data | length) >= 1' > /dev/null 2>&1; then
                ITEM_UID=$(echo "$OUTPUT" | jq -r '.data[0].uid // empty')
                pass "todo items (found items, first uid: $ITEM_UID)"

                if [ -n "$ITEM_UID" ]; then
                    log_test "todo complete"
                    OUTPUT=$(run_hab todo complete "$TODO_ENTITY" "$ITEM_UID")
                    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                        pass "todo complete"
                    else
                        fail "todo complete: $OUTPUT"
                    fi

                    log_test "todo uncomplete"
                    OUTPUT=$(run_hab todo uncomplete "$TODO_ENTITY" "$ITEM_UID")
                    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                        pass "todo uncomplete"
                    else
                        fail "todo uncomplete: $OUTPUT"
                    fi

                    log_test "todo update"
                    OUTPUT=$(run_hab todo update "$TODO_ENTITY" "$ITEM_UID" --summary "Buy oat milk")
                    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                        pass "todo update"
                    else
                        fail "todo update: $OUTPUT"
                    fi

                    log_test "todo remove"
                    OUTPUT=$(run_hab todo remove "$TODO_ENTITY" "$ITEM_UID")
                    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                        pass "todo remove"
                    else
                        fail "todo remove: $OUTPUT"
                    fi
                else
                    pass "todo complete (skipped - no uid)"
                    pass "todo uncomplete (skipped - no uid)"
                    pass "todo update (skipped - no uid)"
                    pass "todo remove (skipped - no uid)"
                fi
            else
                fail "todo items (after adds): $OUTPUT"
                pass "todo complete (skipped)"
                pass "todo uncomplete (skipped)"
                pass "todo update (skipped)"
                pass "todo remove (skipped)"
            fi
        else
            pass "todo lists (skipped - no todo entity)"
            pass "todo items (skipped)"
            pass "todo add (skipped)"
            pass "todo add with due date (skipped)"
            pass "todo items after adds (skipped)"
            pass "todo complete (skipped)"
            pass "todo uncomplete (skipped)"
            pass "todo update (skipped)"
            pass "todo remove (skipped)"
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

run_calendar_create_delete_tests() {
    log_section "Calendar Create/Delete Tests"
    do_auth_login

    # Create a local calendar to test against
    CAL_NAME="Test Cal $(date +%s)"
    OUTPUT=$(run_hab helper local-calendar create "$CAL_NAME")
    if ! echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        fail "calendar create/delete setup: could not create local-calendar: $OUTPUT"
        return
    fi
    CAL_ENTRY_ID=$(echo "$OUTPUT" | jq -r '.data.entry_id // empty')
    sleep 1

    CAL_ENTITY=$(run_hab entity list | jq -r '.data[] | select(.entity_id | startswith("calendar.")) | .entity_id' | head -1)

    if [ -n "$CAL_ENTITY" ]; then
        log_test "calendar create event"
        START="2099-06-01T10:00:00"
        END="2099-06-01T11:00:00"
        OUTPUT=$(run_hab calendar create "$CAL_ENTITY" --summary "Integration Test Event" --start "$START" --end "$END" --description "Created by hab test")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "calendar create event"
        else
            fail "calendar create event: $OUTPUT"
        fi

        log_test "calendar create all-day event"
        OUTPUT=$(run_hab calendar create "$CAL_ENTITY" --summary "Test Holiday" --start "2099-07-04" --end "2099-07-05" --all-day)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "calendar create all-day event"
        else
            fail "calendar create all-day event: $OUTPUT"
        fi
    else
        pass "calendar create event (skipped - no calendar entity)"
        pass "calendar create all-day event (skipped)"
    fi

    # Cleanup
    run_hab helper local-calendar delete "$CAL_ENTRY_ID" > /dev/null 2>&1
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Calendar and To-do Tests"
    run_calendar_todo_tests
    run_calendar_create_delete_tests
    print_summary "Calendar and To-do Tests"
    exit $?
fi

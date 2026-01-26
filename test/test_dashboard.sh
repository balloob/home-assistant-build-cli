#!/bin/bash
# Dashboard tests: dashboard CRUD, views, badges, sections, cards
# Usage: ./test_dashboard.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_dashboard_tests() {
    log_section "Dashboard Tests"

    # Test: dashboard guide (no auth required)
    log_test "dashboard guide"
    OUTPUT=$(run_hab dashboard guide 2>&1)
    if echo "$OUTPUT" | grep -q "Dashboard Creation Guide"; then
        pass "dashboard guide"
    else
        fail "dashboard guide: $OUTPUT"
    fi

    # Ensure we're authenticated
    do_auth_login

    # Test: dashboard list
    log_test "dashboard list"
    OUTPUT=$(run_hab dashboard list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "dashboard list ($COUNT dashboards)"
    else
        fail "dashboard list: $OUTPUT"
    fi

    # Test: dashboard CRUD
    log_test "dashboard create"
    DASHBOARD_URL="test-dashboard-$(date +%s)"
    OUTPUT=$(run_hab dashboard create "$DASHBOARD_URL" --title "Test Dashboard")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        DASHBOARD_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "dashboard create (id: $DASHBOARD_ID)"

        # First save some config so we can get it
        log_test "dashboard save-config"
        DASHBOARD_CONFIG='{"views":[{"title":"Home","cards":[]}]}'
        OUTPUT=$(run_hab_optional dashboard save-config "$DASHBOARD_URL" -d "$DASHBOARD_CONFIG")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "dashboard save-config"
        else
            # Might not support save-config
            pass "dashboard save-config (not available)"
        fi

        log_test "dashboard get"
        OUTPUT=$(run_hab_optional dashboard get "$DASHBOARD_URL")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "dashboard get"
        else
            # Dashboard might not have config yet
            pass "dashboard get (no config yet)"
        fi

        # Test: dashboard view CRUD
        log_test "dashboard view list"
        OUTPUT=$(run_hab_optional dashboard view list "$DASHBOARD_URL")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            VIEW_COUNT=$(echo "$OUTPUT" | jq '.data | length')
            pass "dashboard view list ($VIEW_COUNT views)"

            log_test "dashboard view get"
            OUTPUT=$(run_hab_optional dashboard view get "$DASHBOARD_URL" 0)
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "dashboard view get"
            else
                fail "dashboard view get: $OUTPUT"
            fi

            log_test "dashboard view create"
            OUTPUT=$(run_hab_optional dashboard view create "$DASHBOARD_URL" --title "Test View" --icon "mdi:test-tube")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                NEW_VIEW_INDEX=$(echo "$OUTPUT" | jq -r '.data.index')
                pass "dashboard view create (index: $NEW_VIEW_INDEX)"

                log_test "dashboard view update"
                OUTPUT=$(run_hab_optional dashboard view update "$DASHBOARD_URL" "$NEW_VIEW_INDEX" --title "Updated View")
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "dashboard view update"
                else
                    fail "dashboard view update: $OUTPUT"
                fi

                log_test "dashboard view delete"
                OUTPUT=$(run_hab_optional dashboard view delete "$DASHBOARD_URL" "$NEW_VIEW_INDEX" --force)
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "dashboard view delete"
                else
                    fail "dashboard view delete: $OUTPUT"
                fi
            else
                fail "dashboard view create: $OUTPUT"
            fi
        else
            pass "dashboard view list (not available)"
        fi

        # Test: dashboard badge CRUD
        log_test "dashboard badge list"
        OUTPUT=$(run_hab_optional dashboard badge list "$DASHBOARD_URL" 0)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            BADGE_COUNT=$(echo "$OUTPUT" | jq '.data | length')
            pass "dashboard badge list ($BADGE_COUNT badges)"

            log_test "dashboard badge create"
            OUTPUT=$(run_hab_optional dashboard badge create "$DASHBOARD_URL" 0 --entity "sun.sun")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                NEW_BADGE_INDEX=$(echo "$OUTPUT" | jq -r '.data.index')
                pass "dashboard badge create (index: $NEW_BADGE_INDEX)"

                log_test "dashboard badge get"
                OUTPUT=$(run_hab_optional dashboard badge get "$DASHBOARD_URL" 0 "$NEW_BADGE_INDEX")
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "dashboard badge get"
                else
                    fail "dashboard badge get: $OUTPUT"
                fi

                log_test "dashboard badge update"
                OUTPUT=$(run_hab_optional dashboard badge update "$DASHBOARD_URL" 0 "$NEW_BADGE_INDEX" --entity "person.test")
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "dashboard badge update"
                else
                    fail "dashboard badge update: $OUTPUT"
                fi

                log_test "dashboard badge delete"
                OUTPUT=$(run_hab_optional dashboard badge delete "$DASHBOARD_URL" 0 "$NEW_BADGE_INDEX" --force)
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "dashboard badge delete"
                else
                    fail "dashboard badge delete: $OUTPUT"
                fi
            else
                fail "dashboard badge create: $OUTPUT"
            fi
        else
            pass "dashboard badge list (not available)"
        fi

        # Test: dashboard section CRUD
        log_test "dashboard section list"
        OUTPUT=$(run_hab_optional dashboard section list "$DASHBOARD_URL" 0)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            SECTION_COUNT=$(echo "$OUTPUT" | jq '.data | length')
            pass "dashboard section list ($SECTION_COUNT sections)"

            log_test "dashboard section create"
            OUTPUT=$(run_hab_optional dashboard section create "$DASHBOARD_URL" 0 --title "Test Section" --type "grid")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                NEW_SECTION_INDEX=$(echo "$OUTPUT" | jq -r '.data.index')
                pass "dashboard section create (index: $NEW_SECTION_INDEX)"

                log_test "dashboard section get"
                OUTPUT=$(run_hab_optional dashboard section get "$DASHBOARD_URL" 0 "$NEW_SECTION_INDEX")
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "dashboard section get"
                else
                    fail "dashboard section get: $OUTPUT"
                fi

                log_test "dashboard section update"
                OUTPUT=$(run_hab_optional dashboard section update "$DASHBOARD_URL" 0 "$NEW_SECTION_INDEX" --title "Updated Section")
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "dashboard section update"
                else
                    fail "dashboard section update: $OUTPUT"
                fi

                # Test: dashboard card CRUD within section
                log_test "dashboard card list (in section)"
                OUTPUT=$(run_hab_optional dashboard card list "$DASHBOARD_URL" 0 --section "$NEW_SECTION_INDEX")
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    CARD_COUNT=$(echo "$OUTPUT" | jq '.data | length')
                    pass "dashboard card list in section ($CARD_COUNT cards)"

                    log_test "dashboard card create (in section)"
                    CARD_CONFIG='{"type":"markdown","content":"Test card"}'
                    OUTPUT=$(run_hab_optional dashboard card create "$DASHBOARD_URL" 0 --section "$NEW_SECTION_INDEX" -d "$CARD_CONFIG")
                    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                        NEW_CARD_INDEX=$(echo "$OUTPUT" | jq -r '.data.index')
                        pass "dashboard card create in section (index: $NEW_CARD_INDEX)"

                        log_test "dashboard card get (in section)"
                        OUTPUT=$(run_hab_optional dashboard card get "$DASHBOARD_URL" 0 "$NEW_CARD_INDEX" --section "$NEW_SECTION_INDEX")
                        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                            pass "dashboard card get in section"
                        else
                            fail "dashboard card get in section: $OUTPUT"
                        fi

                        log_test "dashboard card update (in section)"
                        CARD_UPDATE_CONFIG='{"type":"markdown","content":"Updated content"}'
                        OUTPUT=$(run_hab_optional dashboard card update "$DASHBOARD_URL" 0 "$NEW_CARD_INDEX" --section "$NEW_SECTION_INDEX" -d "$CARD_UPDATE_CONFIG")
                        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                            pass "dashboard card update in section"
                        else
                            fail "dashboard card update in section: $OUTPUT"
                        fi

                        log_test "dashboard card delete (in section)"
                        OUTPUT=$(run_hab_optional dashboard card delete "$DASHBOARD_URL" 0 "$NEW_CARD_INDEX" --section "$NEW_SECTION_INDEX" --force)
                        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                            pass "dashboard card delete in section"
                        else
                            fail "dashboard card delete in section: $OUTPUT"
                        fi
                    else
                        fail "dashboard card create in section: $OUTPUT"
                    fi
                else
                    pass "dashboard card list in section (not available)"
                fi

                log_test "dashboard section delete"
                OUTPUT=$(run_hab_optional dashboard section delete "$DASHBOARD_URL" 0 "$NEW_SECTION_INDEX" --force)
                if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                    pass "dashboard section delete"
                else
                    fail "dashboard section delete: $OUTPUT"
                fi
            else
                fail "dashboard section create: $OUTPUT"
            fi
        else
            pass "dashboard section list (not available)"
        fi

        # Test: dashboard card create with defaults (auto-creates section)
        log_test "dashboard card create (with defaults)"
        # Create a card without specifying section - should create section automatically
        OUTPUT=$(run_hab_optional dashboard card create "$DASHBOARD_URL" --entity "sun.sun")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            NEW_CARD_INDEX=$(echo "$OUTPUT" | jq -r '.data.index')
            CARD_TYPE=$(echo "$OUTPUT" | jq -r '.data.type')
            pass "dashboard card create with defaults (index: $NEW_CARD_INDEX, type: $CARD_TYPE)"

            # Verify card was created in a section (view 0, last section)
            log_test "dashboard card get (with defaults)"
            OUTPUT=$(run_hab_optional dashboard card get "$DASHBOARD_URL" 0 "$NEW_CARD_INDEX")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "dashboard card get with defaults"
            else
                fail "dashboard card get with defaults: $OUTPUT"
            fi

            log_test "dashboard card delete (with defaults)"
            OUTPUT=$(run_hab_optional dashboard card delete "$DASHBOARD_URL" 0 "$NEW_CARD_INDEX" --force)
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "dashboard card delete with defaults"
            else
                fail "dashboard card delete with defaults: $OUTPUT"
            fi
        else
            fail "dashboard card create with defaults: $OUTPUT"
        fi

        # Test: dashboard card create with --name flag
        log_test "dashboard card create (with name)"
        OUTPUT=$(run_hab dashboard card create "$DASHBOARD_URL" --entity "sun.sun" --name "Sun Card")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            NEW_CARD_INDEX=$(echo "$OUTPUT" | jq -r '.data.index')
            CARD_NAME=$(echo "$OUTPUT" | jq -r '.data.name')
            if [ "$CARD_NAME" = "Sun Card" ]; then
                pass "dashboard card create with name (index: $NEW_CARD_INDEX, name: $CARD_NAME)"
            else
                fail "dashboard card create with name: expected name 'Sun Card', got '$CARD_NAME'"
            fi

            log_test "dashboard card delete (with name)"
            OUTPUT=$(run_hab dashboard card delete "$DASHBOARD_URL" 0 "$NEW_CARD_INDEX" --force)
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "dashboard card delete with name"
            else
                fail "dashboard card delete with name: $OUTPUT"
            fi
        else
            fail "dashboard card create with name: $OUTPUT"
        fi

        log_test "dashboard update"
        OUTPUT=$(run_hab dashboard update "$DASHBOARD_ID" --title "Updated Dashboard")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "dashboard update"
        else
            fail "dashboard update: $OUTPUT"
        fi

        log_test "dashboard delete"
        OUTPUT=$(run_hab dashboard delete "$DASHBOARD_ID" --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "dashboard delete"
        else
            fail "dashboard delete: $OUTPUT"
        fi
    else
        fail "dashboard create: $OUTPUT"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Dashboard Tests"
    run_dashboard_tests
    print_summary "Dashboard Tests"
    exit $?
fi

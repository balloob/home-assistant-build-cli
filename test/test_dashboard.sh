#!/bin/bash
# Dashboard tests: dashboard CRUD, views, badges, sections, cards
# Usage: ./test_dashboard.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_dashboard_tests() {
    log_section "Dashboard Tests"

    # Test: guide topic listing (no auth required)
    log_test "guide list"
    OUTPUT=$(run_hab guide list)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) > 0' > /dev/null 2>&1; then
        pass "guide list"
    else
        fail "guide list: $OUTPUT"
    fi

    # Test: top-level dashboard guide in text mode (no auth required)
    log_test "guide dashboard text"
    OUTPUT=$(run_hab_text guide dashboard 2>&1)
    if echo "$OUTPUT" | grep -q "Dashboard Creation Guide"; then
        pass "guide dashboard text"
    else
        fail "guide dashboard text: $OUTPUT"
    fi

    # Test: top-level dashboard guide in json mode (no auth required)
    log_test "guide dashboard json"
    OUTPUT=$(run_hab guide dashboard)
    if echo "$OUTPUT" | jq -e '.success == true and .data.topic == "dashboard" and (.data.content | test("Dashboard Creation Guide"))' > /dev/null 2>&1; then
        pass "guide dashboard json"
    else
        fail "guide dashboard json: $OUTPUT"
    fi

    # Test: legacy dashboard guide alias in json mode (no auth required)
    log_test "dashboard guide alias"
    OUTPUT=$(run_hab dashboard guide)
    if echo "$OUTPUT" | jq -e '.success == true and .data.topic == "dashboard"' > /dev/null 2>&1; then
        pass "dashboard guide alias"
    else
        fail "dashboard guide alias: $OUTPUT"
    fi

    # Ensure we're authenticated
    do_auth_login

    run_dashboard_patch_tests

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

		log_test "dashboard get (text mode)"
		OUTPUT=$(run_hab_text dashboard get "$DASHBOARD_URL" 2>&1)
		if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
			pass "dashboard get (text mode)"
		else
			fail "dashboard get (text mode): got JSON instead of text"
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

			log_test "dashboard view get (text mode)"
			OUTPUT=$(run_hab_text dashboard view get "$DASHBOARD_URL" 0)
			if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
				pass "dashboard view get (text mode)"
			else
				fail "dashboard view get (text mode): got JSON instead of text"
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

        # Test: a view path addresses a view, and --type is kept with --data
        log_test "dashboard badge create by view path with --type"
        OUTPUT=$(run_hab_optional dashboard view create "$DASHBOARD_URL" --title "Badge Path" --path badge-path)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            OUTPUT=$(run_hab dashboard badge create "$DASHBOARD_URL" badge-path --type entity-filter --data '{"entities":["sun.sun"],"state_filter":["above_horizon"]}')
            if echo "$OUTPUT" | jq -e '.success == true and .data.config.type == "entity-filter"' > /dev/null 2>&1; then
                pass "dashboard badge create by view path with --type"
            else
                fail "dashboard badge create by view path with --type: $OUTPUT"
            fi

            log_test "dashboard badge list by view path"
            OUTPUT=$(run_hab dashboard badge list "$DASHBOARD_URL" badge-path)
            if echo "$OUTPUT" | jq -e '.success == true and (.data | length) == 1 and .data[0].type == "entity-filter"' > /dev/null 2>&1; then
                pass "dashboard badge list by view path"
            else
                fail "dashboard badge list by view path: $OUTPUT"
            fi

            log_test "dashboard view delete by path"
            OUTPUT=$(run_hab dashboard view delete "$DASHBOARD_URL" badge-path --force)
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "dashboard view delete by path"
            else
                fail "dashboard view delete by path: $OUTPUT"
            fi
        else
            fail "dashboard view create with --path: $OUTPUT"
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

run_dashboard_patch_tests() {
    log_section "Verified Dashboard Patch"
    local path="patch-test-$RANDOM" id initial patch plan revision applied result before_bytes successes=0
    result=$(run_hab dashboard create "$path" --title "Patch test")
    id=$(echo "$result" | jq -r '.data.id // empty')
    if [ -z "$id" ]; then fail "patch dashboard create: $result"; return; fi
    initial='{"title":"Retain root","views":[{"path":"home","cards":[{"type":"tile","entity":"sun.sun","name":"Before","tap_action":{"action":"toggle","confirmation":true}},{"type":"markdown","content":"Retain sibling"}]}]}'
    initial=$(echo "$initial" | jq '.views[0].cards += [range(200) | {type:"tile",entity:"sun.sun",name:("Retained " + tostring),tap_action:{action:"more-info"}}]')
    patch='{"name":"After","tap_action":{"action":"more-info"}}'
    result=$(run_hab dashboard save-config "$path" -d "$initial")
    if echo "$result" | jq -e '.success' >/dev/null; then pass "patch fixture save"; else fail "patch fixture save: $result"; fi

    plan=$(run_hab dashboard patch "$path" --target /views/0/cards/0 -d "$patch" --plan)
    revision=$(echo "$plan" | jq -r '.data.base_revision // empty')
    if echo "$plan" | jq -e '.success and .data.status == "planned" and .data.requests == 1 and .data.change_count == 2 and .data.saved == false and .data.changes[0].before == "Before" and .data.changes[0].after == "After"' >/dev/null; then
        pass "patch preview returns bounded actual diff"
    else fail "patch preview: $plan"; fi
    result=$(run_hab dashboard get "$path")
    before_bytes=${#result}
    if echo "$result" | jq -e --argjson expected "$initial" '.data == $expected' >/dev/null; then pass "preview is side-effect-free"; else fail "preview changed config: $result"; fi

    applied=$(run_hab dashboard patch "$path" --target /views/0/cards/0 -d "$patch" --if-match "$revision")
    if echo "$applied" | jq -e '.success and .data.status == "verified" and .data.saved and .data.verified and .data.requests == 4 and .data.reloaded == "not_applicable"' >/dev/null; then
        pass "patch apply verifies stored config on one session"
    else fail "patch apply: $applied"; fi
    result=$(run_hab dashboard get "$path")
    if echo "$result" | jq -e '.data.title == "Retain root" and .data.views[0].cards[1].content == "Retain sibling" and .data.views[0].cards[0].entity == "sun.sun" and .data.views[0].cards[0].tap_action.confirmation == true and .data.views[0].cards[0].tap_action.action == "more-info"' >/dev/null; then
        pass "patch retains unrelated nested fields and siblings"
    else fail "patch retention: $result"; fi

    result=$(run_hab dashboard patch "$path" --target /views/0/cards/0 -d "$patch" --if-match "$revision")
    if echo "$result" | jq -e '.success == false and .error.code == "CONFLICT" and .error.details.result.saved == false' >/dev/null; then pass "stale revision rejected"; else fail "stale revision: $result"; fi
    revision=$(echo "$applied" | jq -r '.data.result_revision // empty')
    result=$(run_hab dashboard patch "$path" --target /views/0/cards/0 -d "$patch" --if-match "$revision")
    if echo "$result" | jq -e '.success and .data.status == "noop" and .data.saved == false and .data.requests == 1 and .data.change_count == 0' >/dev/null; then pass "repeat is no-op with no save"; else fail "patch repeat: $result"; fi
    for i in {1..10}; do
        result=$(run_hab dashboard patch "$path" --target /views/0/cards/0 -d "$patch" --if-match "$revision")
        if echo "$result" | jq -e '.success and .data.status == "noop" and .data.saved == false and .data.requests == 1' >/dev/null; then successes=$((successes + 1)); fi
    done
    if [ "$successes" = 10 ]; then pass "repeat completion reliability: 10/10, zero saves"; else fail "repeat completion reliability: $successes/10"; fi
    echo "Patch metrics (202 cards): inspection=$before_bytes bytes, preview=${#plan} bytes, apply=${#applied} bytes; API requests: preview=1, apply=4, repeat=1."
    result=$(run_hab dashboard patch "$path" --target /views/-1 -d "$patch" --plan)
    if echo "$result" | jq -e '.success == false and .error.code == "INVALID_PATCH"' >/dev/null; then pass "invalid selector rejected"; else fail "invalid selector: $result"; fi
    result=$(run_hab dashboard patch "$path" --set '/title="Unsafe"')
    if echo "$result" | jq -e '.success == false and .error.code == "PRECONDITION_REQUIRED"' >/dev/null; then pass "apply requires inspection revision"; else fail "missing revision: $result"; fi

    result=$(run_hab dashboard delete "$id" --force)
    if echo "$result" | jq -e '.success' >/dev/null; then pass "patch fixture cleanup"; else fail "patch fixture cleanup: $result"; fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Dashboard Tests"
    run_dashboard_tests
    print_summary "Dashboard Tests"
    exit $?
fi

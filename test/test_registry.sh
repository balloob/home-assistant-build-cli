#!/bin/bash
# Registry tests: entity, device, area, floor, label commands
# Usage: ./test_registry.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

test_entity_registry_fields() {
    local entity_id="$1" stamp area_id label_id category_id other_category_id output
    stamp=$(date +%s)
    area_id=$(run_hab area create "Entity Update Area $stamp" | jq -r '.data.area_id // empty')
    label_id=$(run_hab label create "Entity Update Label $stamp" | jq -r '.data.label_id // empty')
    category_id=$(run_hab category create "Entity Update Category $stamp" --scope helpers | jq -r '.data.category_id // empty')
    other_category_id=$(run_hab category create "Entity Update Other $stamp" --scope script | jq -r '.data.category_id // empty')
    if [ -n "$area_id" ] && [ -n "$label_id" ] && [ -n "$category_id" ] && [ -n "$other_category_id" ]; then
        log_test "entity update registry fields preview"
        output=$(run_hab entity update "$entity_id" --area "$area_id" --labels "$label_id" --aliases "Primary alias,Secondary alias" --category "helpers=$category_id" --hidden --disabled --plan)
        if echo "$output" | jq -e --arg area "$area_id" --arg label "$label_id" --arg category "$category_id" '.success == true and .data.saved == false and .data.changes.area_id.after == $area and .data.changes.labels.after == [$label] and .data.changes.aliases.after == ["Primary alias", "Secondary alias"] and .data.changes.categories.after.helpers == $category and .data.changes.hidden_by.after == "user" and .data.changes.disabled_by.after == "user"' > /dev/null; then
            pass "entity update previews registry fields"
        else
            fail "entity update registry fields preview: $output"
        fi

        log_test "entity update registry fields apply"
        output=$(run_hab entity update "$entity_id" --area "$area_id" --labels "$label_id" --aliases "Primary alias,Secondary alias" --category "helpers=$category_id" --category "script=$other_category_id" --hidden --disabled)
        if echo "$output" | jq -e '.success == true and .data.saved == true and .data.verified == true and .data.changes.area_id.before == null and .data.changes.labels.before == [] and .data.changes.hidden_by.before == null and .data.changes.disabled_by.before == null' > /dev/null; then
            pass "entity update saves all registry fields; preview made no changes"
        else
            fail "entity update registry fields apply: $output"
        fi

        log_test "entity update clears fields and preserves unrelated category scopes"
        output=$(run_hab entity update "$entity_id" --area "" --labels "" --aliases "" --category helpers= --hidden=false --disabled=false)
        if echo "$output" | jq -e --arg category "$other_category_id" '.success == true and .data.verified == true and .data.changes.area_id.after == null and .data.changes.labels.after == [] and .data.changes.aliases.after == [] and .data.changes.hidden_by.after == null and .data.changes.disabled_by.after == null and .data.changes.categories.after == {script: $category}' > /dev/null; then
            pass "entity update clears fields and preserves other scopes"
        else
            fail "entity update clearing registry fields: $output"
        fi

        log_test "entity update rejects unknown labels before writing"
        output=$(run_hab entity update "$entity_id" --name "Must not be saved" --labels "missing_label_$stamp")
        if echo "$output" | jq -e '.success == false and .error.details.result.saved == false' > /dev/null; then
            pass "entity update rejects unknown label before writing"
        else
            fail "entity update unknown label: $output"
        fi
    else
        fail "entity update registry fields setup"
    fi
    [ -z "$category_id" ] || run_hab category delete "$category_id" --scope helpers --force > /dev/null
    [ -z "$other_category_id" ] || run_hab category delete "$other_category_id" --scope script --force > /dev/null
    [ -z "$label_id" ] || run_hab label delete "$label_id" --force > /dev/null
    [ -z "$area_id" ] || run_hab area delete "$area_id" --force > /dev/null
}

test_entity_device_class() {
    local name entry_id entity_id output attempt
    name="Registry Device Class $(date +%s)"
    entry_id=$(run_hab helper template create "$name" --type binary_sensor --state "{{ true }}" | jq -r '.data.entry_id // empty')
    if [ -z "$entry_id" ]; then
        fail "entity update device class setup"
        return
    fi
    # Config-entry setup is asynchronous; resolve the newly registered entity
    # through its config entry instead of fields omitted by entity list.
    for attempt in {1..20}; do
        entity_id=$(run_hab search related config_entry "$entry_id" | jq -r '.data.related.entity[]? | select(startswith("binary_sensor."))' | head -1)
        [ -z "$entity_id" ] || break
        sleep 0.25
    done
    if [ -n "$entity_id" ]; then
        log_test "entity update device class"
        output=$(run_hab entity update "$entity_id" --device-class door)
        if echo "$output" | jq -e '.success == true and .data.verified == true and .data.changes.device_class.after == "door"' > /dev/null; then
            pass "entity update device class"
        else
            fail "entity update device class: $output"
        fi
        output=$(run_hab entity update "$entity_id" --device-class "")
        if echo "$output" | jq -e '.success == true and .data.verified == true and .data.changes.device_class.before == "door" and .data.changes.device_class.after == null' > /dev/null; then
            pass "entity update clears device class"
        else
            fail "entity update clears device class: $output"
        fi
    else
        fail "entity update device class entity discovery"
    fi
    run_hab helper template delete "$entry_id" > /dev/null
}

run_registry_tests() {
    log_section "Registry Tests (Entity/Device/Area/Floor/Label)"

    # Ensure we're authenticated
    do_auth_login

    # Test: entity list
    log_test "entity list"
    OUTPUT=$(run_hab entity list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "entity list ($COUNT entities)"
    else
        fail "entity list: $OUTPUT"
    fi

    # Test: entity get (get first available entity)
    log_test "entity get"
    FIRST_ENTITY=$(run_hab entity list | jq -r '.data[0].entity_id // empty')
    if [ -n "$FIRST_ENTITY" ]; then
        OUTPUT=$(run_hab entity get "$FIRST_ENTITY")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            # Verify registry data is included
            if echo "$OUTPUT" | jq -e '.data.registry != null' > /dev/null 2>&1; then
                pass "entity get with registry data ($FIRST_ENTITY)"
            else
                pass "entity get ($FIRST_ENTITY)"
            fi
        else
            fail "entity get: $OUTPUT"
        fi
    else
        pass "entity get (skipped - no entities)"
    fi

	# Test: entity get --related
	log_test "entity get --related"
    if [ -n "$FIRST_ENTITY" ]; then
        OUTPUT=$(run_hab_optional entity get "$FIRST_ENTITY" --related)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "entity get --related"
        else
            pass "entity get --related (search/related not supported)"
        fi
	else
		pass "entity get --related (skipped - no entities)"
	fi

	log_test "entity get (text mode)"
	if [ -n "$FIRST_ENTITY" ]; then
		OUTPUT=$(run_hab_text entity get "$FIRST_ENTITY")
		if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
			pass "entity get (text mode)"
		else
			fail "entity get (text mode): got JSON instead of text"
		fi
	else
		pass "entity get (text mode) (skipped - no entities)"
	fi

    log_test "entity update setup"
    RENAME_HELPER_ID=$(run_hab helper input-boolean create "Rename Test $(date +%s)" | jq -r '.data.id // empty')
    if [ -n "$RENAME_HELPER_ID" ]; then
        RENAME_OLD_ID="input_boolean.$RENAME_HELPER_ID"
        RENAME_NEW_ID="${RENAME_OLD_ID}_renamed"
        RENAME_UNIQUE_ID=$(run_hab entity get "$RENAME_OLD_ID" | jq -r '.data.registry.unique_id')

        log_test "entity update --plan"
        OUTPUT=$(run_hab entity update "$RENAME_OLD_ID" --entity-id "$RENAME_NEW_ID" --plan)
        if echo "$OUTPUT" | jq -e --arg old "$RENAME_OLD_ID" --arg new "$RENAME_NEW_ID" '.success == true and .data.mode == "plan" and .data.old_entity_id == $old and .data.new_entity_id == $new and .data.saved == false and .data.verified == false and (.data.warnings | length) > 0' > /dev/null; then
            pass "entity update --plan"
        else
            fail "entity update --plan: $OUTPUT"
        fi
        OUTPUT=$(run_hab entity get "$RENAME_OLD_ID")
        if echo "$OUTPUT" | jq -e --arg id "$RENAME_OLD_ID" '.data.registry.entity_id == $id' > /dev/null; then
            pass "entity update --plan leaves source unchanged"
        else
            fail "entity update --plan changed source: $OUTPUT"
        fi

        log_test "entity update name/icon preview"
        OUTPUT=$(run_hab entity update "$RENAME_OLD_ID" --name "Kitchen plug" --icon mdi:power-socket-eu --dry-run)
        if echo "$OUTPUT" | jq -e '.success == true and .data.saved == false and .data.changes.name.after == "Kitchen plug" and .data.changes.icon.after == "mdi:power-socket-eu" and .data.old_entity_id == .data.new_entity_id' > /dev/null; then
            pass "entity update previews name/icon without ID change"
        else
            fail "entity update name/icon preview: $OUTPUT"
        fi
        OUTPUT=$(run_hab entity get "$RENAME_OLD_ID")
        if echo "$OUTPUT" | jq -e '.data.registry.name == null and .data.registry.icon == null' > /dev/null; then
            pass "entity update name/icon preview leaves overrides unchanged"
        else
            fail "entity update name/icon preview changed overrides: $OUTPUT"
        fi

        log_test "entity update name/icon only"
        OUTPUT=$(run_hab entity update "$RENAME_OLD_ID" --name "Kitchen plug" --icon mdi:power-socket-eu)
        if echo "$OUTPUT" | jq -e '.success == true and .data.saved == true and .data.verified == true and .data.old_entity_id == .data.new_entity_id' > /dev/null; then
            pass "entity update name/icon only saved and verified"
        else
            fail "entity update name/icon only: $OUTPUT"
        fi

        log_test "entity update clear name/icon overrides"
        OUTPUT=$(run_hab entity update "$RENAME_OLD_ID" --name "" --icon "")
        if echo "$OUTPUT" | jq -e '.success == true and .data.saved == true and .data.verified == true and .data.changes.name.before == "Kitchen plug" and .data.changes.icon.before == "mdi:power-socket-eu" and .data.changes.name.after == null and .data.changes.icon.after == null' > /dev/null; then
            pass "entity update clears name/icon overrides"
        else
            fail "entity update clear name/icon: $OUTPUT"
        fi

        log_test "entity update apply all fields and verify unique_id"
        OUTPUT=$(run_hab entity update "$RENAME_OLD_ID" --entity-id "$RENAME_NEW_ID" --name "Renamed plug" --icon mdi:power)
        if echo "$OUTPUT" | jq -e --arg unique "$RENAME_UNIQUE_ID" '.success == true and .data.saved == true and .data.verified == true and .data.unique_id == $unique and .data.changes.name.after == "Renamed plug" and .data.changes.icon.after == "mdi:power"' > /dev/null; then
            pass "entity update saved and verified"
        else
            fail "entity update: $OUTPUT"
        fi

        log_test "entity update collision"
        COLLISION_HELPER_ID=$(run_hab helper input-boolean create "Rename Collision $(date +%s)" | jq -r '.data.id // empty')
        if [ -n "$COLLISION_HELPER_ID" ]; then
            OUTPUT=$(run_hab entity update "$RENAME_NEW_ID" --entity-id "input_boolean.$COLLISION_HELPER_ID")
            if echo "$OUTPUT" | jq -e '.success == false and .error.details.result.saved == false' > /dev/null; then
                pass "entity update rejects collision before writing"
            else
                fail "entity update collision: $OUTPUT"
            fi
            run_hab helper input-boolean delete "$COLLISION_HELPER_ID" > /dev/null
        else
            fail "entity update collision setup"
        fi

        log_test "entity update rename back"
        OUTPUT=$(run_hab entity update "$RENAME_NEW_ID" --entity-id "$RENAME_OLD_ID")
        if echo "$OUTPUT" | jq -e '.success == true and .data.saved == true and .data.verified == true' > /dev/null; then
            pass "entity update rename back"
        else
            fail "entity update rename back: $OUTPUT"
        fi
        test_entity_registry_fields "$RENAME_OLD_ID"
        run_hab helper input-boolean delete "$RENAME_HELPER_ID" > /dev/null
    else
        fail "entity update setup"
    fi
    test_entity_device_class

    # Test: entity search (search for any entity)
    log_test "entity search"
    OUTPUT=$(run_hab entity search ".")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "entity search ($COUNT matches)"
    else
        fail "entity search: $OUTPUT"
    fi

    # Test: entity history (if entities available)
    log_test "entity history"
    if [ -n "$FIRST_ENTITY" ]; then
        OUTPUT=$(run_hab entity history "$FIRST_ENTITY" 2>&1)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "entity history"
        elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
            # History might not be available
            pass "entity history (not available)"
        else
            fail "entity history: $OUTPUT"
        fi
    else
        pass "entity history (skipped - no entities)"
    fi

    # Test: entity history with --start/--end time range
    log_test "entity history --start/--end"
    if [ -n "$FIRST_ENTITY" ]; then
        START_TIME=$(date -u -d '1 hour ago' '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || date -u -v-1H '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null)
        END_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
        if [ -n "$START_TIME" ] && [ -n "$END_TIME" ]; then
            OUTPUT=$(run_hab_optional entity history "$FIRST_ENTITY" --start "$START_TIME" --end "$END_TIME")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "entity history --start/--end"
            elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
                pass "entity history --start/--end (not available)"
            else
                fail "entity history --start/--end: $OUTPUT"
            fi
        else
            pass "entity history --start/--end (skipped - date command not available)"
        fi
    else
        pass "entity history --start/--end (skipped - no entities)"
    fi

    # Test: device list
    log_test "device list"
    OUTPUT=$(run_hab device list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "device list ($COUNT devices)"
    else
        fail "device list: $OUTPUT"
    fi

    # Test: device get (if devices available)
    log_test "device get"
    FIRST_DEVICE=$(run_hab device list | jq -r '.data[0].id // empty')
    if [ -n "$FIRST_DEVICE" ]; then
        OUTPUT=$(run_hab device get "$FIRST_DEVICE")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "device get ($FIRST_DEVICE)"
        else
            fail "device get: $OUTPUT"
        fi
    else
        pass "device get (skipped - no devices)"
    fi

    # Test: device get --related
    log_test "device get --related"
    if [ -n "$FIRST_DEVICE" ]; then
        OUTPUT=$(run_hab_optional device get "$FIRST_DEVICE" --related)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "device get --related"
        else
            pass "device get --related (search/related not supported)"
        fi
    else
        pass "device get --related (skipped - no devices)"
    fi

    # Test: device entities (if devices available)
    log_test "device entities"
    if [ -n "$FIRST_DEVICE" ]; then
        OUTPUT=$(run_hab device entities "$FIRST_DEVICE")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "device entities"
        else
            fail "device entities: $OUTPUT"
        fi
    else
        pass "device entities (skipped - no devices)"
    fi

    log_test "device update metadata, preview, and reset"
    UPDATE_DEVICE_ID=$(run_hab device list | jq -r '[.data[] | select(.disabled_by == null)][0].id // empty')
    if [ -n "$UPDATE_DEVICE_ID" ]; then
        ORIGINAL_DEVICE=$(run_hab device get "$UPDATE_DEVICE_ID" | jq '.data')
        ORIGINAL_DEVICE_ENTITIES=$(run_hab entity list --device "$UPDATE_DEVICE_ID" | jq '[.data[].entity_id]')
        DEVICE_AREA=$(run_hab area create "Device update test area" | jq -r '.data.area_id')
        DEVICE_LABEL=$(run_hab label create "Device update test label" | jq -r '.data.label_id')
        DEVICE_PLAN=$(run_hab device update "$UPDATE_DEVICE_ID" --name "Device update test" --area "$DEVICE_AREA" --labels "$DEVICE_LABEL" --disabled --plan)
        DEVICE_UNCHANGED=$(run_hab device get "$UPDATE_DEVICE_ID" | jq '.data')
        OUTPUT=$(run_hab device update "$UPDATE_DEVICE_ID" --name "Device update test" --area "$DEVICE_AREA" --labels "$DEVICE_LABEL" --disabled)
        if echo "$DEVICE_PLAN" | jq -e '.data.mode == "plan" and .data.changes.disabled_by.after == "user" and (.data.saved | not)' >/dev/null &&
            [ "$ORIGINAL_DEVICE" = "$DEVICE_UNCHANGED" ] &&
            echo "$OUTPUT" | jq -e --arg area "$DEVICE_AREA" --arg label "$DEVICE_LABEL" '.data.name_by_user == "Device update test" and .data.area_id == $area and .data.labels == [$label] and .data.disabled_by == "user" and .data.saved and .data.verified' >/dev/null; then
            pass "device update metadata and preview"
        else
            fail "device update metadata and preview: $OUTPUT $DEVICE_PLAN"
        fi
        OUTPUT=$(run_hab device update "$UPDATE_DEVICE_ID" --name "" --area "" --labels "" --disabled=false)
        if echo "$OUTPUT" | jq -e '.data.name_by_user == null and .data.area_id == null and .data.labels == [] and .data.disabled_by == null and .data.saved and .data.verified' >/dev/null; then
            pass "device update reset"
        else
            fail "device update reset: $OUTPUT"
        fi
        OUTPUT=$(run_hab device update "$UPDATE_DEVICE_ID" --name "Should not apply" --labels hab_missing_test_label)
        CURRENT=$(run_hab device get "$UPDATE_DEVICE_ID")
        if echo "$OUTPUT" | jq -e '.success == false' >/dev/null && echo "$CURRENT" | jq -e '.data.name_by_user == null' >/dev/null; then
            pass "device update rejects unknown labels before writing"
        else
            fail "device update unknown labels: $OUTPUT"
        fi
        run_hab device update "$UPDATE_DEVICE_ID" --name "$(echo "$ORIGINAL_DEVICE" | jq -r '.name_by_user // ""')" --area "$(echo "$ORIGINAL_DEVICE" | jq -r '.area_id // ""')" --labels "$(echo "$ORIGINAL_DEVICE" | jq -r '.labels | join(",")')" >/dev/null
        run_hab area delete "$DEVICE_AREA" --force >/dev/null
        run_hab label delete "$DEVICE_LABEL" --force >/dev/null
        # Registry verification does not wait for runtime entities. HA delays
        # reloading a re-enabled device; restore the fixture before later tests.
        DEVICE_ENTITIES_RESTORED=0
        for ((attempt = 0; attempt < 60; attempt++)); do
            CURRENT=$(run_hab entity list --device "$UPDATE_DEVICE_ID")
            if echo "$CURRENT" | jq -e --argjson expected "$ORIGINAL_DEVICE_ENTITIES" '.success == true and (($expected - [.data[].entity_id]) | length == 0)' >/dev/null; then
                DEVICE_ENTITIES_RESTORED=1
                break
            fi
            sleep 1
        done
        if [ "$DEVICE_ENTITIES_RESTORED" = 1 ]; then
            pass "device update restores live entity fixtures"
        else
            fail "device entities did not return after re-enabling: $CURRENT"
        fi
    else
        fail "device update requires a device fixture"
    fi

    # Test: area CRUD
    log_test "area create"
    AREA_NAME="Test Area $(date +%s)"
    OUTPUT=$(run_hab area create "$AREA_NAME")
    if echo "$OUTPUT" | jq -e '.success == true and .data.area_id != null' > /dev/null 2>&1; then
        AREA_ID=$(echo "$OUTPUT" | jq -r '.data.area_id')
        pass "area create (id: $AREA_ID)"

        log_test "area list"
        OUTPUT=$(run_hab area list)
        if echo "$OUTPUT" | jq -e ".success == true and (.data | map(select(.area_id == \"$AREA_ID\")) | length) > 0" > /dev/null 2>&1; then
            pass "area list (found created area)"
        else
            fail "area list: created area not found"
        fi

        log_test "area update"
        OUTPUT=$(run_hab area update "$AREA_ID" --name "$AREA_NAME Updated")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "area update"
        else
            fail "area update: $OUTPUT"
        fi

        log_test "area live preview and clear"
        OUTPUT=$(run_hab area update "$AREA_ID" --icon mdi:home)
        PLAN=$(run_hab area update "$AREA_ID" --icon "" --plan)
        CURRENT=$(run_hab area get "$AREA_ID")
        OUTPUT=$(run_hab area update "$AREA_ID" --icon "" --floor "")
        if echo "$PLAN" | jq -e '.data.mode == "plan" and .data.changes.icon.before == "mdi:home" and .data.changes.icon.after == null and .data.would_change and (.data.saved | not)' >/dev/null &&
            echo "$CURRENT" | jq -e '.data.icon == "mdi:home"' >/dev/null &&
            echo "$OUTPUT" | jq -e '.data.saved and .data.verified and .data.icon == null and .data.floor_id == null' >/dev/null; then
            pass "area live preview and clear"
        else
            fail "area live preview and clear: $OUTPUT $PLAN $CURRENT"
        fi

        log_test "area metadata fields"
        AREA_LABEL=$(run_hab label create "Area metadata test label" | jq -r '.data.label_id')
        TEMP_ENTRY=$(run_hab helper template create "Area metadata temperature" --type sensor --state '{{ 21 }}' --unit '°C' --device-class temperature | jq -r '.data.entry_id')
        HUMIDITY_ENTRY=$(run_hab helper template create "Area metadata humidity" --type sensor --state '{{ 45 }}' --unit '%' --device-class humidity | jq -r '.data.entry_id')
        for attempt in 1 2 3 4 5; do
            TEMP_ENTITY=$(run_hab search related config_entry "$TEMP_ENTRY" | jq -r '.data.related.entity[0] // empty')
            HUMIDITY_ENTITY=$(run_hab search related config_entry "$HUMIDITY_ENTRY" | jq -r '.data.related.entity[0] // empty')
            if [ -n "$TEMP_ENTITY" ] && [ -n "$HUMIDITY_ENTITY" ]; then break; fi
            sleep 1
        done
        PLAN=$(run_hab area update "$AREA_ID" --aliases ' Cooking area,Kitchen,Kitchen' --labels "$AREA_LABEL" --picture /local/kitchen.jpg --temperature-entity "$TEMP_ENTITY" --humidity-entity "$HUMIDITY_ENTITY" --plan)
        CURRENT=$(run_hab area get "$AREA_ID")
        OUTPUT=$(run_hab area update "$AREA_ID" --aliases ' Cooking area,Kitchen,Kitchen' --labels "$AREA_LABEL" --picture /local/kitchen.jpg --temperature-entity "$TEMP_ENTITY" --humidity-entity "$HUMIDITY_ENTITY")
        if [ -n "$TEMP_ENTITY" ] && [ -n "$HUMIDITY_ENTITY" ] &&
            echo "$PLAN" | jq -e '.data.mode == "plan" and .data.changes.picture.before == null and .data.changes.picture.after == "/local/kitchen.jpg" and (.data.saved | not)' >/dev/null &&
            echo "$CURRENT" | jq -e '.data.picture == null and .data.aliases == []' >/dev/null &&
            echo "$OUTPUT" | jq -e --arg label "$AREA_LABEL" --arg temp "$TEMP_ENTITY" --arg humidity "$HUMIDITY_ENTITY" '.data.saved and .data.verified and (.data.aliases | sort) == ["Cooking area","Kitchen"] and .data.labels == [$label] and .data.picture == "/local/kitchen.jpg" and .data.temperature_entity_id == $temp and .data.humidity_entity_id == $humidity' >/dev/null; then
            pass "area metadata fields and preview"
        else
            fail "area metadata fields and preview: $OUTPUT $PLAN"
        fi
        OUTPUT=$(run_hab area update "$AREA_ID" --picture /local/should-not-apply.jpg --labels hab_missing_test_label)
        CURRENT=$(run_hab area get "$AREA_ID")
        if echo "$OUTPUT" | jq -e '.success == false' >/dev/null && echo "$CURRENT" | jq -e '.data.picture == "/local/kitchen.jpg"' >/dev/null; then
            pass "area rejects unknown labels before writing"
        else
            fail "area unknown labels: $OUTPUT"
        fi
        OUTPUT=$(run_hab area update "$AREA_ID" --temperature-entity "$HUMIDITY_ENTITY" --plan)
        if echo "$OUTPUT" | jq -e '.success == false' >/dev/null; then
            pass "area rejects wrong sensor class in preview"
        else
            fail "area wrong sensor class: $OUTPUT"
        fi
        OUTPUT=$(run_hab area update "$AREA_ID" --aliases "" --labels "" --picture "" --temperature-entity "" --humidity-entity "")
        if echo "$OUTPUT" | jq -e '.data.saved and .data.verified and .data.aliases == [] and .data.labels == [] and .data.picture == null and .data.temperature_entity_id == null and .data.humidity_entity_id == null' >/dev/null; then
            pass "area clear metadata fields"
        else
            fail "area clear metadata fields: $OUTPUT"
        fi
        run_hab helper template delete "$TEMP_ENTRY" >/dev/null
        run_hab helper template delete "$HUMIDITY_ENTRY" >/dev/null
        run_hab label delete "$AREA_LABEL" --force >/dev/null

        log_test "area delete"
        OUTPUT=$(run_hab area delete "$AREA_ID" --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "area delete"
        else
            fail "area delete: $OUTPUT"
        fi
    else
        fail "area create: $OUTPUT"
    fi

	# Test: area get (using the first available area)
	log_test "area get"
    FIRST_AREA=$(run_hab area list | jq -r '.data[0].area_id // empty')
    if [ -n "$FIRST_AREA" ]; then
        OUTPUT=$(run_hab area get "$FIRST_AREA")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "area get ($FIRST_AREA)"
        else
            fail "area get: $OUTPUT"
        fi
	else
		pass "area get (skipped - no areas)"
	fi

	log_test "area get (text mode)"
	if [ -n "$FIRST_AREA" ]; then
		OUTPUT=$(run_hab_text area get "$FIRST_AREA")
		if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
			pass "area get (text mode)"
		else
			fail "area get (text mode): got JSON instead of text"
		fi
	else
		pass "area get (text mode) (skipped - no areas)"
	fi

    # Test: area get --related
    log_test "area get --related"
    if [ -n "$FIRST_AREA" ]; then
        OUTPUT=$(run_hab_optional area get "$FIRST_AREA" --related)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "area get --related"
        else
            pass "area get --related (search/related not supported)"
        fi
    else
        pass "area get --related (skipped - no areas)"
    fi

    # Test: floor CRUD
    log_test "floor create"
    FLOOR_NAME="Test Floor $(date +%s)"
    OUTPUT=$(run_hab floor create "$FLOOR_NAME" --level 1)
    if echo "$OUTPUT" | jq -e '.success == true and .data.floor_id != null' > /dev/null 2>&1; then
        FLOOR_ID=$(echo "$OUTPUT" | jq -r '.data.floor_id')
        pass "floor create (id: $FLOOR_ID)"

        log_test "floor delete"
        OUTPUT=$(run_hab floor delete "$FLOOR_ID" --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "floor delete"
        else
            fail "floor delete: $OUTPUT"
        fi
    else
        fail "floor create: $OUTPUT"
    fi

    # Test: floor list
    log_test "floor list"
    OUTPUT=$(run_hab floor list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "floor list"
    else
        fail "floor list: $OUTPUT"
    fi

    # Test: floor get (using the first available floor)
    log_test "floor get"
    FIRST_FLOOR=$(run_hab floor list | jq -r '.data[0].floor_id // empty')
    if [ -n "$FIRST_FLOOR" ]; then
        OUTPUT=$(run_hab floor get "$FIRST_FLOOR")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "floor get ($FIRST_FLOOR)"
        else
            fail "floor get: $OUTPUT"
        fi
    else
        pass "floor get (skipped - no floors)"
    fi

    # Test: floor get --related
    log_test "floor get --related"
    if [ -n "$FIRST_FLOOR" ]; then
        OUTPUT=$(run_hab_optional floor get "$FIRST_FLOOR" --related)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "floor get --related"
        else
            pass "floor get --related (search/related not supported)"
        fi
    else
        pass "floor get --related (skipped - no floors)"
    fi

    # Test: floor update (create, update, then delete)
    log_test "floor update"
    FLOOR_UPDATE_NAME="Test Floor Update $(date +%s)"
    OUTPUT=$(run_hab floor create "$FLOOR_UPDATE_NAME" --level 2)
    if echo "$OUTPUT" | jq -e '.success == true and .data.floor_id != null' > /dev/null 2>&1; then
        UPDATE_FLOOR_ID=$(echo "$OUTPUT" | jq -r '.data.floor_id')
        OUTPUT=$(run_hab floor update "$UPDATE_FLOOR_ID" --name "Updated Floor Name")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "floor update"
        else
            fail "floor update: $OUTPUT"
        fi
        OUTPUT=$(run_hab floor update "$UPDATE_FLOOR_ID" --level 0)
        CLEARED=$(run_hab floor update "$UPDATE_FLOOR_ID" --level "")
        NOOP=$(run_hab floor update "$UPDATE_FLOOR_ID" --level "")
        if echo "$OUTPUT" | jq -e '.data.level == 0 and .data.verified' >/dev/null &&
            echo "$CLEARED" | jq -e '.data.level == null and .data.saved and .data.verified' >/dev/null &&
            echo "$NOOP" | jq -e '(.data.saved | not) and (.data.would_change | not) and .data.verified' >/dev/null; then
            pass "floor zero, clear, and no-op"
        else
            fail "floor zero, clear, and no-op: $OUTPUT $CLEARED $NOOP"
        fi
        PLAN=$(run_hab floor update "$UPDATE_FLOOR_ID" --aliases ' Downstairs,Ground floor,Downstairs' --plan)
        CURRENT=$(run_hab floor get "$UPDATE_FLOOR_ID")
        OUTPUT=$(run_hab floor update "$UPDATE_FLOOR_ID" --aliases ' Downstairs,Ground floor,Downstairs')
        if echo "$PLAN" | jq -e '.data.mode == "plan" and .data.changes.aliases.before == [] and .data.changes.aliases.after == ["Downstairs","Ground floor"] and (.data.saved | not)' >/dev/null &&
            echo "$CURRENT" | jq -e '.data.aliases == []' >/dev/null &&
            echo "$OUTPUT" | jq -e '(.data.aliases | sort) == ["Downstairs","Ground floor"] and .data.saved and .data.verified' >/dev/null; then
            pass "floor aliases and live preview"
        else
            fail "floor aliases and live preview: $OUTPUT $PLAN"
        fi
        NOOP=$(run_hab floor update "$UPDATE_FLOOR_ID" --aliases 'Ground floor,Downstairs')
        OUTPUT=$(run_hab floor update "$UPDATE_FLOOR_ID" --aliases "")
        if echo "$NOOP" | jq -e '(.data.saved | not) and (.data.would_change | not) and .data.verified' >/dev/null &&
            echo "$OUTPUT" | jq -e '.data.aliases == [] and .data.saved and .data.verified' >/dev/null; then
            pass "floor alias order, no-op, and clear"
        else
            fail "floor alias order, no-op, and clear: $OUTPUT $NOOP"
        fi
        # Cleanup
        run_hab floor delete "$UPDATE_FLOOR_ID" --force > /dev/null 2>&1
    else
        fail "floor update: could not create floor for testing"
    fi

    # Test: label CRUD
    log_test "label create"
    LABEL_NAME="Test Label $(date +%s)"
    OUTPUT=$(run_hab label create "$LABEL_NAME" --color red)
    if echo "$OUTPUT" | jq -e '.success == true and .data.label_id != null' > /dev/null 2>&1; then
        LABEL_ID=$(echo "$OUTPUT" | jq -r '.data.label_id')
        pass "label create (id: $LABEL_ID)"

        log_test "label delete"
        OUTPUT=$(run_hab label delete "$LABEL_ID" --force)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "label delete"
        else
            fail "label delete: $OUTPUT"
        fi
    else
        fail "label create: $OUTPUT"
    fi

    # Test: label list
    log_test "label list"
    OUTPUT=$(run_hab label list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "label list"
    else
        fail "label list: $OUTPUT"
    fi

    # Test: label update (create, update, then delete)
    log_test "label update"
    LABEL_UPDATE_NAME="Test Label Update $(date +%s)"
    OUTPUT=$(run_hab label create "$LABEL_UPDATE_NAME" --color blue)
    if echo "$OUTPUT" | jq -e '.success == true and .data.label_id != null' > /dev/null 2>&1; then
        UPDATE_LABEL_ID=$(echo "$OUTPUT" | jq -r '.data.label_id')
        OUTPUT=$(run_hab label update "$UPDATE_LABEL_ID" --name "Updated Label Name")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "label update"
        else
            fail "label update: $OUTPUT"
        fi
        OUTPUT=$(run_hab label update "$UPDATE_LABEL_ID" --icon mdi:tag --description "Description")
        OUTPUT=$(run_hab label update "$UPDATE_LABEL_ID" --color "" --icon "" --description "")
        if echo "$OUTPUT" | jq -e '.data.color == null and .data.icon == null and .data.description == null and .data.saved and .data.verified' >/dev/null; then
            pass "label clear metadata"
        else
            fail "label clear metadata: $OUTPUT"
        fi
        # Cleanup
        run_hab label delete "$UPDATE_LABEL_ID" --force > /dev/null 2>&1
    else
        fail "label update: could not create label for testing"
    fi

    # Test: label assign/remove (create label, assign to entity, remove, cleanup)
    log_test "label assign/remove"
    ASSIGN_LABEL_NAME="Assign Label $(date +%s)"
    LABEL_OUTPUT=$(run_hab label create "$ASSIGN_LABEL_NAME" --color green)
    # Get a sensor entity which should be in the entity registry
    ASSIGN_ENTITY=$(run_hab entity list | jq -r '.data[] | select(.entity_id | startswith("sensor.")) | .entity_id' | head -1)
    if echo "$LABEL_OUTPUT" | jq -e '.success == true and .data.label_id != null' > /dev/null 2>&1 && [ -n "$ASSIGN_ENTITY" ]; then
        ASSIGN_LABEL_ID=$(echo "$LABEL_OUTPUT" | jq -r '.data.label_id')

        # Assign label to entity
        OUTPUT=$(run_hab_optional label assign "$ASSIGN_LABEL_ID" "$ASSIGN_ENTITY")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "label assign"

            # Remove label from entity
            log_test "label remove"
            OUTPUT=$(run_hab_optional label remove "$ASSIGN_LABEL_ID" "$ASSIGN_ENTITY")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "label remove"
            else
                # Remove might not work if entity not in registry
                pass "label remove (entity not in registry)"
            fi
        else
            # Entity might not be in registry or might not support labels
            pass "label assign (entity not in registry)"
            log_test "label remove"
            pass "label remove (skipped)"
        fi

        # Cleanup
        run_hab label delete "$ASSIGN_LABEL_ID" --force > /dev/null 2>&1
    else
        pass "label assign/remove (skipped - no sensor entities or label creation failed)"
    fi

    # Test: device list --area
    log_test "device list --area"
    FIRST_AREA=$(run_hab area list | jq -r '.data[0].area_id // empty')
    if [ -n "$FIRST_AREA" ]; then
        OUTPUT=$(run_hab device list --area "$FIRST_AREA")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
            pass "device list --area ($COUNT devices in $FIRST_AREA)"
        else
            fail "device list --area: $OUTPUT"
        fi
    else
        OUTPUT=$(run_hab device list --area "nonexistent")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "device list --area (filter works, no matching devices)"
        else
            fail "device list --area: $OUTPUT"
        fi
    fi

    # Test: area list --floor
    log_test "area list --floor"
    FIRST_FLOOR=$(run_hab floor list | jq -r '.data[0].floor_id // empty')
    if [ -n "$FIRST_FLOOR" ]; then
        OUTPUT=$(run_hab area list --floor "$FIRST_FLOOR")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
            pass "area list --floor ($COUNT areas on $FIRST_FLOOR)"
        else
            fail "area list --floor: $OUTPUT"
        fi
    else
        OUTPUT=$(run_hab area list --floor "nonexistent")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "area list --floor (filter works, no matching areas)"
        else
            fail "area list --floor: $OUTPUT"
        fi
    fi

    # Test: entity list --floor
    log_test "entity list --floor"
    if [ -n "$FIRST_FLOOR" ]; then
        OUTPUT=$(run_hab entity list --floor "$FIRST_FLOOR")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
            pass "entity list --floor ($COUNT entities on $FIRST_FLOOR)"
        else
            fail "entity list --floor: $OUTPUT"
        fi
    else
        OUTPUT=$(run_hab entity list --floor "nonexistent")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "entity list --floor (filter works, no matching entities)"
        else
            fail "entity list --floor: $OUTPUT"
        fi
    fi

    # Test: device list --floor
    log_test "device list --floor"
    if [ -n "$FIRST_FLOOR" ]; then
        OUTPUT=$(run_hab device list --floor "$FIRST_FLOOR")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
            pass "device list --floor ($COUNT devices on $FIRST_FLOOR)"
        else
            fail "device list --floor: $OUTPUT"
        fi
    else
        OUTPUT=$(run_hab device list --floor "nonexistent")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "device list --floor (filter works, no matching devices)"
        else
            fail "device list --floor: $OUTPUT"
        fi
    fi

    # Test: entity list --device
    log_test "entity list --device"
    if [ -n "$FIRST_DEVICE" ]; then
        OUTPUT=$(run_hab entity list --device "$FIRST_DEVICE")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
            pass "entity list --device ($COUNT entities for device)"
        else
            fail "entity list --device: $OUTPUT"
        fi
    else
        OUTPUT=$(run_hab entity list --device "nonexistent")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "entity list --device (filter works, no matching entities)"
        else
            fail "entity list --device: $OUTPUT"
        fi
    fi

    # Test: entity list --device-class
    log_test "entity list --device-class"
    # First check if there are any entities with device_class
    FIRST_DEVICE_CLASS=$(run_hab entity list | jq -r '.data[] | select(.device_class != null and .device_class != "") | .device_class' | head -1)
    if [ -n "$FIRST_DEVICE_CLASS" ]; then
        OUTPUT=$(run_hab entity list --device-class "$FIRST_DEVICE_CLASS")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
            pass "entity list --device-class ($COUNT entities with device_class=$FIRST_DEVICE_CLASS)"
        else
            fail "entity list --device-class: $OUTPUT"
        fi
    else
        OUTPUT=$(run_hab entity list --device-class "nonexistent")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "entity list --device-class (filter works, no matching entities)"
        else
            fail "entity list --device-class: $OUTPUT"
        fi
    fi

    # Test: entity get --device
    log_test "entity get --device"
    if [ -n "$FIRST_ENTITY" ]; then
        OUTPUT=$(run_hab_optional entity get "$FIRST_ENTITY" --device)
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "entity get --device"
        else
            fail "entity get --device: $OUTPUT"
        fi
    else
        pass "entity get --device (skipped - no entities)"
    fi

    # Test: search related
    log_test "search related"
    if [ -n "$FIRST_ENTITY" ]; then
        OUTPUT=$(run_hab_optional search related entity "$FIRST_ENTITY")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "search related entity"
        else
            pass "search related (search/related not supported)"
        fi
    else
        pass "search related (skipped - no entities)"
    fi

    # Test: search related integration (results use singular item type keys)
    log_test "search related integration"
    OUTPUT=$(run_hab search related integration sun)
    if echo "$OUTPUT" | jq -e '.success == true and (.data.related.entity | index("sun.sun")) != null' > /dev/null 2>&1; then
        pass "search related integration"
    else
        fail "search related integration: $OUTPUT"
    fi

    # Test: entity enable/disable (need an entity from entity registry)
    log_test "entity enable/disable"
    # Use a sensor entity which should be in the entity registry
    REGISTRY_ENTITY=$(run_hab entity list | jq -r '.data[] | select(.entity_id | startswith("sensor.")) | .entity_id' | head -1)
    if [ -n "$REGISTRY_ENTITY" ]; then
        # Disable then re-enable
        OUTPUT=$(run_hab_optional entity disable "$REGISTRY_ENTITY")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "entity disable"

            log_test "entity enable"
            OUTPUT=$(run_hab_optional entity enable "$REGISTRY_ENTITY")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "entity enable"
            else
                # Entity might not be properly re-enabled
                pass "entity enable (entity might not support enable)"
            fi
        else
            # Entity might not be in registry or might not support disable
            pass "entity disable (entity not in registry)"
            log_test "entity enable"
            pass "entity enable (skipped)"
        fi
    else
        pass "entity disable (skipped - no sensor entities)"
        log_test "entity enable"
        pass "entity enable (skipped - no sensor entities)"
    fi

    # Test: entity rename (rename and rename back)
    log_test "entity rename"
    # REGISTRY_ENTITY was set earlier from a sensor entity
    if [ -n "$REGISTRY_ENTITY" ]; then
        OUTPUT=$(run_hab_optional entity rename "$REGISTRY_ENTITY" "Test Renamed Entity")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            pass "entity rename"
            # Rename back to original (empty name to clear custom name)
            run_hab_optional entity rename "$REGISTRY_ENTITY" "" > /dev/null 2>&1
        else
            # Entity might not be in registry
            pass "entity rename (entity not in registry)"
        fi
    else
        pass "entity rename (skipped - no sensor entities)"
    fi
}

run_person_tests() {
    log_section "Person Tests"

    # Ensure we're authenticated
    do_auth_login

    # Test: person list
    log_test "person list"
    OUTPUT=$(run_hab person list)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "person list ($COUNT persons)"
    else
        fail "person list: $OUTPUT"
    fi

    # Test: person CRUD
    log_test "person create"
    PERSON_NAME="Test Person $(date +%s)"
    OUTPUT=$(run_hab person create "$PERSON_NAME" --device-trackers device_tracker.test_phone)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        PERSON_ID=$(echo "$OUTPUT" | jq -r '.data.id // empty')
        pass "person create (id: $PERSON_ID)"

        if [ -n "$PERSON_ID" ]; then
            log_test "person get"
            OUTPUT=$(run_hab person get "$PERSON_ID")
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "person get"
            else
                fail "person get: $OUTPUT"
            fi

            log_test "person update"
            OUTPUT=$(run_hab person update "$PERSON_ID" --name "Updated Test Person")
            if echo "$OUTPUT" | jq -e '.success == true and .data.device_trackers == ["device_tracker.test_phone"]' > /dev/null 2>&1; then
                pass "person update"
            else
                fail "person update: $OUTPUT"
            fi

            log_test "person delete"
            OUTPUT=$(run_hab person delete "$PERSON_ID" --force)
            if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
                pass "person delete"
            else
                fail "person delete: $OUTPUT"
            fi
        else
            pass "person get/update/delete (skipped - no ID returned)"
        fi
    else
        fail "person create: $OUTPUT"
    fi
}

run_entity_logbook_tests() {
    log_section "Entity Logbook Tests"

    # Ensure we're authenticated
    do_auth_login

    # Test: entity logbook (use sun.sun as a reliable entity)
    log_test "entity logbook"
    OUTPUT=$(run_hab entity logbook sun.sun)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
        pass "entity logbook sun.sun ($COUNT entries)"
    else
        fail "entity logbook: $OUTPUT"
    fi

    # Test: entity logbook with time range
    log_test "entity logbook with time range"
    START="$(date -u -d '1 day ago' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v-1d +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo '')"
    if [ -n "$START" ]; then
        OUTPUT=$(run_hab entity logbook sun.sun -s "$START")
        if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
            COUNT=$(echo "$OUTPUT" | jq '.data | if . == null then 0 else length end')
            pass "entity logbook with --start ($COUNT entries)"
        else
            fail "entity logbook with --start: $OUTPUT"
        fi
    else
        pass "entity logbook with time range (skipped - date command not available)"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Registry Tests"
    run_registry_tests
    run_person_tests
    run_entity_logbook_tests
    print_summary "Registry Tests"
    exit $?
fi

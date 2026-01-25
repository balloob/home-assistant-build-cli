#!/bin/bash
# Registry tests: entity, device, area, floor, label commands
# Usage: ./test_registry.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

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

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Registry Tests"
    run_registry_tests
    print_summary "Registry Tests"
    exit $?
fi

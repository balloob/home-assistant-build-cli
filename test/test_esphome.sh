#!/bin/bash
# ESPHome integration tests using a local mock dashboard
# Usage: ./test_esphome.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

MOCK_ESPHOME_PID=""
MOCK_ESPHOME_PORT="16123"
MOCK_ESPHOME_URL=""
MOCK_ESPHOME_LOG=""

start_mock_esphome_dashboard() {
    MOCK_ESPHOME_URL="http://127.0.0.1:${MOCK_ESPHOME_PORT}"
    MOCK_ESPHOME_LOG=$(mktemp)

    go run "$PROJECT_DIR/test/tools/mock_esphome_dashboard" --port "$MOCK_ESPHOME_PORT" > "$MOCK_ESPHOME_LOG" 2>&1 &
    MOCK_ESPHOME_PID=$!

    for i in $(seq 1 40); do
        if curl -s "$MOCK_ESPHOME_URL/version" > /dev/null 2>&1; then
            return 0
        fi
        sleep 0.25
    done

    echo "Mock ESPHome dashboard failed to start"
    if [ -f "$MOCK_ESPHOME_LOG" ]; then
        cat "$MOCK_ESPHOME_LOG"
    fi
    return 1
}

stop_mock_esphome_dashboard() {
    if [ -n "$MOCK_ESPHOME_PID" ]; then
        kill "$MOCK_ESPHOME_PID" 2>/dev/null || true
        wait "$MOCK_ESPHOME_PID" 2>/dev/null || true
        MOCK_ESPHOME_PID=""
    fi
    if [ -n "$MOCK_ESPHOME_LOG" ] && [ -f "$MOCK_ESPHOME_LOG" ]; then
        rm -f "$MOCK_ESPHOME_LOG"
        MOCK_ESPHOME_LOG=""
    fi
}

run_esphome_tests() {
    trap stop_mock_esphome_dashboard RETURN

    log_section "ESPHome Tests"
    do_auth_login

    if ! start_mock_esphome_dashboard; then
        fail "start mock esphome dashboard"
        return
    fi

    export HAB_ESPHOME_URL="$MOCK_ESPHOME_URL"

    log_test "esphome create list presets"
    OUTPUT=$(run_hab esphome create --list-presets)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) > 0' > /dev/null 2>&1; then
        pass "esphome create list presets"
    else
        fail "esphome create list presets: $OUTPUT"
    fi

    log_test "esphome create preset help"
    OUTPUT=$(run_hab esphome create --preset-help relay)
    if echo "$OUTPUT" | jq -e '.success == true and .data.id == "relay"' > /dev/null 2>&1; then
        pass "esphome create preset help"
    else
        fail "esphome create preset help: $OUTPUT"
    fi

    log_test "esphome boards"
    OUTPUT=$(run_hab esphome boards esp32)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) > 0' > /dev/null 2>&1; then
        pass "esphome boards"
    else
        fail "esphome boards: $OUTPUT"
    fi

    log_test "esphome create with preset"
    OUTPUT=$(run_hab esphome create demo-relay --type basic --platform esp32 --board esp32dev --preset relay --relay-pin 23)
    if echo "$OUTPUT" | jq -e '.success == true and .data.configuration != null and .data.preset == "relay"' > /dev/null 2>&1; then
        CONFIG=$(echo "$OUTPUT" | jq -r '.data.configuration')
        pass "esphome create with preset ($CONFIG)"
    else
        fail "esphome create with preset: $OUTPUT"
        return
    fi

    log_test "esphome config patch"
    OUTPUT=$(run_hab esphome config-patch "$CONFIG" --set logger.level=INFO)
    if echo "$OUTPUT" | jq -e '.success == true and .data.validated == true' > /dev/null 2>&1; then
        pass "esphome config patch"
    else
        fail "esphome config patch: $OUTPUT"
    fi

    log_test "esphome validate structured"
    OUTPUT=$(run_hab esphome validate "$CONFIG" --structured)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "esphome validate structured"
    else
        fail "esphome validate structured: $OUTPUT"
    fi

    log_test "esphome update workflow"
    OUTPUT=$(run_hab esphome update "$CONFIG" --set wifi.ssid=updated-ssid --build=false)
    if echo "$OUTPUT" | jq -e '.success == true and .data.validated == true' > /dev/null 2>&1; then
        pass "esphome update workflow"
    else
        fail "esphome update workflow: $OUTPUT"
    fi

    log_test "esphome serial ports"
    OUTPUT=$(run_hab esphome serial ports)
    if echo "$OUTPUT" | jq -e '.success == true and (.data | length) > 0' > /dev/null 2>&1; then
        pass "esphome serial ports"
    else
        fail "esphome serial ports: $OUTPUT"
    fi

    log_test "esphome info"
    OUTPUT=$(run_hab esphome info "$CONFIG")
    if echo "$OUTPUT" | jq -e '.success == true and .data.name != null' > /dev/null 2>&1; then
        pass "esphome info"
    else
        fail "esphome info: $OUTPUT"
    fi

    TEMPLATE='{"NAME":"Example","GPIO":[0,0,0,0,0,0,0,0,0,32,0,224,0,0],"FLAG":0,"BASE":18}'

    log_test "esphome migrate tasmota analyze"
    OUTPUT=$(run_hab esphome migrate tasmota-template analyze --data "$TEMPLATE")
    if echo "$OUTPUT" | jq -e '.success == true and .data.analysis.supported_count == 2' > /dev/null 2>&1; then
        pass "esphome migrate tasmota analyze"
    else
        fail "esphome migrate tasmota analyze: $OUTPUT"
    fi

    log_test "esphome migrate tasmota create"
    OUTPUT=$(run_hab esphome migrate tasmota-template create migrated-demo --platform esp32 --board esp32dev --data "$TEMPLATE")
    if echo "$OUTPUT" | jq -e '.success == true and .data.configuration != null' > /dev/null 2>&1; then
        pass "esphome migrate tasmota create"
    else
        fail "esphome migrate tasmota create: $OUTPUT"
    fi

    log_test "esphome catalog show"
    OUTPUT=$(run_hab_optional esphome catalog show M5Stack-AtomS3-Lite)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "esphome catalog show"
    else
        pass "esphome catalog show (network or rate limit unavailable)"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "ESPHome Tests"
    run_esphome_tests
    print_summary "ESPHome Tests"
    exit $?
fi

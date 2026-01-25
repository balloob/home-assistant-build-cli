#!/bin/bash
# Common test library for hab CLI integration tests
# Source this file from individual test files

# Set strict mode
set -e

# Directory setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_DIR="$(dirname "$TEST_DIR")"
HAB="$PROJECT_DIR/hab"

# Test token from empty-hass
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIyZWZkZGJjZmY0MzQ0NGRlYmUyMDhkNDUyM2RlNTIwMSIsImlhdCI6MTc2OTI4MjkyNiwiZXhwIjoyMDg0NjQyOTI2fQ.ZYSmLdcv5EfGCXrwO2Nd6bxHrxxU-7ieuE0ySwurU9A"
URL="http://localhost:8124"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters (exported for use across sourced files)
export PASSED=0
export FAILED=0

# Track if we started empty-hass (for cleanup)
export EMPTY_HASS_PID=""

# Config directory - can be set externally for suite runs
if [ -z "$HAB_TEST_CONFIG_DIR" ]; then
    export HAB_TEST_CONFIG_DIR=$(mktemp -d)
    export HAB_TEST_OWN_CONFIG=1
else
    export HAB_TEST_OWN_CONFIG=0
fi

# Check if we should manage empty-hass ourselves
# Set HAB_TEST_HASS_RUNNING=1 when running as part of a suite
if [ -z "$HAB_TEST_HASS_RUNNING" ]; then
    export HAB_TEST_MANAGE_HASS=1
else
    export HAB_TEST_MANAGE_HASS=0
fi

cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    if [ "$HAB_TEST_MANAGE_HASS" = "1" ] && [ -n "$EMPTY_HASS_PID" ]; then
        kill $EMPTY_HASS_PID 2>/dev/null || true
        wait $EMPTY_HASS_PID 2>/dev/null || true
    fi
    if [ "$HAB_TEST_OWN_CONFIG" = "1" ] && [ -n "$HAB_TEST_CONFIG_DIR" ]; then
        rm -rf "$HAB_TEST_CONFIG_DIR"
    fi
    echo -e "${YELLOW}Done.${NC}"
}

log_test() {
    echo -e "\n${YELLOW}TEST: $1${NC}"
}

log_section() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

pass() {
    echo -e "${GREEN}PASS${NC}: $1"
    PASSED=$((PASSED + 1))
}

fail() {
    echo -e "${RED}FAIL${NC}: $1"
    FAILED=$((FAILED + 1))
}

run_hab() {
    "$HAB" --config "$HAB_TEST_CONFIG_DIR" "$@"
}

# Run command that might fail or not be supported
run_hab_optional() {
    "$HAB" --config "$HAB_TEST_CONFIG_DIR" "$@" 2>&1 || echo '{"success":false,"error":{"message":"command failed"}}'
}

wait_for_hass() {
    echo "Waiting for Home Assistant to be ready..."
    for i in {1..60}; do
        STATE=$(curl -s -H "Authorization: Bearer $TOKEN" "$URL/api/config" 2>/dev/null | jq -r '.state // empty')
        if [ "$STATE" = "RUNNING" ]; then
            echo "Home Assistant is ready (state: RUNNING)!"
            return 0
        elif [ -n "$STATE" ]; then
            echo "Home Assistant state: $STATE (waiting for RUNNING)..."
        fi
        sleep 2
    done
    echo "Home Assistant did not become ready in time"
    return 1
}

build_hab() {
    echo -e "${YELLOW}Building hab CLI...${NC}"
    cd "$PROJECT_DIR"
    go build -o hab .
    echo "Built: $HAB"
}

start_empty_hass() {
    if [ "$HAB_TEST_MANAGE_HASS" = "1" ]; then
        echo -e "\n${YELLOW}Starting empty-hass...${NC}"
        uvx --from git+https://github.com/balloob/empty-hass empty-hass --port 8124 > /dev/null 2>&1 &
        EMPTY_HASS_PID=$!
        echo "Started empty-hass with PID: $EMPTY_HASS_PID"
        wait_for_hass
    else
        echo -e "${YELLOW}Using existing empty-hass instance...${NC}"
    fi
}

# Authenticate with Home Assistant
do_auth_login() {
    run_hab auth login --token --url "$URL" --access-token "$TOKEN" > /dev/null 2>&1
}

# Print test summary and return appropriate exit code
print_summary() {
    local test_name="${1:-Tests}"
    echo -e "\n${YELLOW}========================================${NC}"
    echo -e "${YELLOW}${test_name} Summary${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${GREEN}Passed: $PASSED${NC}"
    echo -e "${RED}Failed: $FAILED${NC}"

    if [ $FAILED -eq 0 ]; then
        echo -e "\n${GREEN}All tests passed!${NC}"
        return 0
    else
        echo -e "\n${RED}Some tests failed.${NC}"
        return 1
    fi
}

# Initialize test environment for standalone runs
init_standalone_test() {
    local test_name="$1"
    trap cleanup EXIT
    build_hab
    start_empty_hass
    echo -e "\n${YELLOW}Running ${test_name}...${NC}"
    echo "Config directory: $HAB_TEST_CONFIG_DIR"
}

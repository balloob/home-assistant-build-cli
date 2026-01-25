#!/bin/bash
# Integration test orchestrator for hab CLI
# Runs all test files with a single empty-hass instance
#
# Usage:
#   ./run_integration_test.sh           # Run all tests
#   ./run_integration_test.sh core      # Run only core tests
#   ./run_integration_test.sh registry  # Run only registry tests
#   ./run_integration_test.sh automation # Run only automation tests
#   ./run_integration_test.sh script    # Run only script tests
#   ./run_integration_test.sh dashboard # Run only dashboard tests
#   ./run_integration_test.sh misc      # Run only misc tests

set -e

# Save test directory before common.sh overwrites SCRIPT_DIR
TEST_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source common library
source "$TEST_ROOT/lib/common.sh"

# Cleanup function specific to orchestrator
orchestrator_cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    if [ -n "$EMPTY_HASS_PID" ]; then
        kill $EMPTY_HASS_PID 2>/dev/null || true
        wait $EMPTY_HASS_PID 2>/dev/null || true
    fi
    if [ -n "$HAB_TEST_CONFIG_DIR" ]; then
        rm -rf "$HAB_TEST_CONFIG_DIR"
    fi
    echo -e "${YELLOW}Done.${NC}"
}

trap orchestrator_cleanup EXIT

# Build the CLI
build_hab

# Create shared config directory
export HAB_TEST_CONFIG_DIR=$(mktemp -d)
export HAB_TEST_OWN_CONFIG=0

# Start empty-hass (we manage it, not the individual tests)
export HAB_TEST_MANAGE_HASS=0
echo -e "\n${YELLOW}Starting empty-hass...${NC}"
uvx --from git+https://github.com/balloob/empty-hass empty-hass --port 8124 > /dev/null 2>&1 &
EMPTY_HASS_PID=$!
echo "Started empty-hass with PID: $EMPTY_HASS_PID"

# Wait for it to be ready
wait_for_hass

# Mark that hass is running for test files
export HAB_TEST_HASS_RUNNING=1

echo -e "\n${YELLOW}Running tests...${NC}"
echo "Config directory: $HAB_TEST_CONFIG_DIR"

# Determine which tests to run
TEST_FILTER="${1:-all}"

run_test_file() {
    local test_file="$1"
    local test_name="$2"

    if [ -f "$test_file" ]; then
        source "$test_file"
        "run_${test_name}_tests"
    else
        echo -e "${RED}Test file not found: $test_file${NC}"
        return 1
    fi
}

case "$TEST_FILTER" in
    all)
        run_test_file "$TEST_ROOT/test_core.sh" "core"
        run_test_file "$TEST_ROOT/test_registry.sh" "registry"
        run_test_file "$TEST_ROOT/test_automation.sh" "automation"
        run_test_file "$TEST_ROOT/test_script.sh" "script"
        run_test_file "$TEST_ROOT/test_dashboard.sh" "dashboard"
        run_test_file "$TEST_ROOT/test_misc.sh" "misc"
        ;;
    core)
        run_test_file "$TEST_ROOT/test_core.sh" "core"
        ;;
    registry)
        # Need to authenticate first since core tests aren't running
        do_auth_login
        run_test_file "$TEST_ROOT/test_registry.sh" "registry"
        ;;
    automation)
        do_auth_login
        run_test_file "$TEST_ROOT/test_automation.sh" "automation"
        ;;
    script)
        do_auth_login
        run_test_file "$TEST_ROOT/test_script.sh" "script"
        ;;
    dashboard)
        do_auth_login
        run_test_file "$TEST_ROOT/test_dashboard.sh" "dashboard"
        ;;
    misc)
        do_auth_login
        run_test_file "$TEST_ROOT/test_misc.sh" "misc"
        ;;
    *)
        echo -e "${RED}Unknown test filter: $TEST_FILTER${NC}"
        echo "Usage: $0 [all|core|registry|automation|script|dashboard|misc]"
        exit 1
        ;;
esac

# Print summary
print_summary "Integration Tests"
exit $?

#!/bin/bash
# Integration test orchestrator for hab CLI using empty-hass
# Usage: ./run_integration_test.sh [test_group...]
#
# Test groups: core, registry, automation, script, dashboard, helpers, template, calendar, misc
# Run all tests: ./run_integration_test.sh (no arguments)
# Run specific tests: ./run_integration_test.sh core registry

set -e

# Store orchestrator script directory before sourcing common.sh (which redefines SCRIPT_DIR)
ORCHESTRATOR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source common test library
source "$ORCHESTRATOR_DIR/lib/common.sh"

# Set trap for cleanup
trap cleanup EXIT

# Available test groups and their corresponding functions
declare -A TEST_GROUPS=(
    ["core"]="run_core_tests"
    ["registry"]="run_registry_tests"
    ["automation"]="run_automation_tests"
    ["script"]="run_script_tests"
    ["dashboard"]="run_dashboard_tests"
    ["helpers"]="run_helpers_tests"
    ["template"]="run_template_tests"
    ["calendar"]="run_calendar_todo_tests"
    ["misc"]="run_misc_tests"
)

# Order of test execution (matters for dependencies)
TEST_ORDER=(core registry automation script dashboard helpers template calendar misc)

# Source all test files
source_test_files() {
    source "$ORCHESTRATOR_DIR/test_core.sh"
    source "$ORCHESTRATOR_DIR/test_registry.sh"
    source "$ORCHESTRATOR_DIR/test_automation.sh"
    source "$ORCHESTRATOR_DIR/test_script.sh"
    source "$ORCHESTRATOR_DIR/test_dashboard.sh"
    source "$ORCHESTRATOR_DIR/test_helpers.sh"
    source "$ORCHESTRATOR_DIR/test_template.sh"
    source "$ORCHESTRATOR_DIR/test_calendar_todo.sh"
    source "$ORCHESTRATOR_DIR/test_misc.sh"
}

# Print usage
print_usage() {
    echo "Usage: $0 [test_group...]"
    echo ""
    echo "Available test groups:"
    echo "  core       - Auth and system tests"
    echo "  registry   - Entity, device, area, floor, label tests"
    echo "  automation - Automation CRUD and trigger/condition/action tests"
    echo "  script     - Script CRUD and script-action tests"
    echo "  dashboard  - Dashboard, views, badges, sections, cards tests"
    echo "  helpers    - Helper types (input_boolean, counter, timer, group, etc.)"
    echo "  template   - Template entity tests"
    echo "  calendar   - Calendar and to-do list tests"
    echo "  misc       - Actions, zones, backups, blueprints tests"
    echo ""
    echo "Examples:"
    echo "  $0              # Run all tests"
    echo "  $0 core         # Run only core tests"
    echo "  $0 core registry # Run core and registry tests"
}

# Run selected tests
run_tests() {
    local groups=("$@")

    # If no groups specified, run all
    if [ ${#groups[@]} -eq 0 ]; then
        groups=("${TEST_ORDER[@]}")
    fi

    # Validate test groups
    for group in "${groups[@]}"; do
        if [ "$group" = "-h" ] || [ "$group" = "--help" ]; then
            print_usage
            exit 0
        fi
        if [ -z "${TEST_GROUPS[$group]}" ]; then
            echo -e "${RED}Unknown test group: $group${NC}"
            print_usage
            exit 1
        fi
    done

    # Run each test group
    for group in "${groups[@]}"; do
        local func="${TEST_GROUPS[$group]}"
        echo -e "\n${BLUE}Running test group: $group${NC}"
        $func
    done
}

# Main entry point
main() {
    # Build CLI
    build_hab

    # Start empty-hass
    start_empty_hass

    echo -e "\n${YELLOW}Running integration tests...${NC}"
    echo "Config directory: $HAB_TEST_CONFIG_DIR"

    # Mark that we're running as orchestrator so test files don't start their own empty-hass
    export HAB_TEST_HASS_RUNNING=1

    # Source all test files
    source_test_files

    # Run selected tests
    run_tests "$@"

    # Print summary
    print_summary "Integration Tests"
}

# Run main with all arguments
main "$@"

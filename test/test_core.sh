#!/bin/bash
# Core tests: auth and system commands
# Usage: ./test_core.sh (standalone) or source from run_integration_test.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

run_core_tests() {
    log_section "Core Tests (Auth & System)"

    # Test: auth login
    log_test "auth login"
    OUTPUT=$(run_hab auth login --token --url "$URL" --access-token "$TOKEN")
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "auth login"
    else
        fail "auth login: $OUTPUT"
    fi

    # Test: auth status
    log_test "auth status"
    OUTPUT=$(run_hab auth status)
    if echo "$OUTPUT" | jq -e '.success == true and .data.authenticated == true' > /dev/null 2>&1; then
        pass "auth status"
    else
        fail "auth status: $OUTPUT"
    fi

    # Test: system info
    log_test "system info"
    OUTPUT=$(run_hab system info)
    if echo "$OUTPUT" | jq -e '.success == true and .data.version != null' > /dev/null 2>&1; then
        VERSION=$(echo "$OUTPUT" | jq -r '.data.version')
        pass "system info (version: $VERSION)"
    else
        fail "system info: $OUTPUT"
    fi

    # Test: system health
    log_test "system health"
    OUTPUT=$(run_hab system health)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "system health"
    else
        fail "system health: $OUTPUT"
    fi

    # Test: text output mode (now the default)
    log_test "text output mode"
    OUTPUT=$(run_hab_text system info)
    if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
        # Not valid JSON, which is what we expect for text mode
        if echo "$OUTPUT" | grep -qi "version"; then
            pass "text output mode"
        else
            fail "text output mode: expected version info"
        fi
    else
        fail "text output mode: got JSON instead of text"
    fi

    # Test: system config check (may not work with empty-hass)
    log_test "system config check"
    OUTPUT=$(run_hab_optional system config-check)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "system config check"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "system config check (not available)"
    else
        fail "system config check: $OUTPUT"
    fi

    # Test: system updates (may not work with empty-hass)
    log_test "system updates"
    OUTPUT=$(run_hab_optional system updates)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "system updates"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "system updates (not available)"
    else
        fail "system updates: $OUTPUT"
    fi

    # Test: system logs (may not work with empty-hass)
    log_test "system logs"
    OUTPUT=$(run_hab_optional system logs)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "system logs"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        pass "system logs (not available)"
    else
        fail "system logs: $OUTPUT"
    fi

    # Test: auth refresh (should fail if using token auth, but tests the command works)
    log_test "auth refresh"
    # First login again for this test
    run_hab auth login --token --url "$URL" --access-token "$TOKEN" > /dev/null 2>&1
    OUTPUT=$(run_hab_optional auth refresh)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "auth refresh"
    elif echo "$OUTPUT" | jq -e '.success == false' > /dev/null 2>&1; then
        # Expected for token auth - refresh not needed
        pass "auth refresh (not needed for token auth)"
    else
        # Command might return error text for non-OAuth auth
        if echo "$OUTPUT" | grep -qi "oauth\|token"; then
            pass "auth refresh (correctly requires OAuth)"
        else
            fail "auth refresh: $OUTPUT"
        fi
    fi

    # Test: auth logout
    log_test "auth logout"
    OUTPUT=$(run_hab auth logout)
    if echo "$OUTPUT" | jq -e '.success == true' > /dev/null 2>&1; then
        pass "auth logout"
    else
        fail "auth logout: $OUTPUT"
    fi

    # Test: auth status after logout
    log_test "auth status after logout"
    OUTPUT=$(run_hab auth status)
    if echo "$OUTPUT" | jq -e '.success == true and .data.authenticated == false' > /dev/null 2>&1; then
        pass "auth status after logout"
    else
        fail "auth status after logout: $OUTPUT"
    fi

    # Re-login for other tests that may run after
    do_auth_login

    # ==========================================================================
    # Version command tests
    # ==========================================================================
    log_test "version (JSON)"
    OUTPUT=$(run_hab version)
    if echo "$OUTPUT" | jq -e '.success == true and .data.version != null and .data.os != null and .data.arch != null' > /dev/null 2>&1; then
        VERSION=$(echo "$OUTPUT" | jq -r '.data.version')
        pass "version JSON (version: $VERSION)"
    else
        fail "version JSON: $OUTPUT"
    fi

    log_test "version (text)"
    OUTPUT=$(run_hab_text version)
    if echo "$OUTPUT" | grep -q "version"; then
        pass "version text"
    else
        fail "version text: $OUTPUT"
    fi

    # ==========================================================================
    # Text output mode tests (beyond system info)
    # ==========================================================================
    log_test "text mode: auth status"
    OUTPUT=$(run_hab_text auth status)
    if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
        # Not JSON — good, it's text mode
        if echo "$OUTPUT" | grep -qi "authenticated\|url\|auth"; then
            pass "text mode: auth status"
        else
            fail "text mode: auth status (unexpected output)"
        fi
    else
        fail "text mode: auth status (got JSON instead of text)"
    fi

    log_test "text mode: entity list"
    OUTPUT=$(run_hab_text entity list)
    if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
        pass "text mode: entity list"
    else
        fail "text mode: entity list (got JSON instead of text)"
    fi

    log_test "text mode: area list"
    OUTPUT=$(run_hab_text area list)
    if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
        pass "text mode: area list"
    else
        fail "text mode: area list (got JSON instead of text)"
    fi

    log_test "text mode: action list"
    OUTPUT=$(run_hab_text action list)
    if ! echo "$OUTPUT" | jq . > /dev/null 2>&1; then
        pass "text mode: action list"
    else
        fail "text mode: action list (got JSON instead of text)"
    fi

    # ==========================================================================
    # Verbose flag test
    # ==========================================================================
    log_test "verbose flag"
    OUTPUT=$("$HAB" --config "$HAB_TEST_CONFIG_DIR" --json --verbose version 2>&1)
    if echo "$OUTPUT" | grep -q "level=debug"; then
        # Verbose produces debug log lines AND the normal output
        if echo "$OUTPUT" | grep -q '"success"'; then
            pass "verbose flag (debug logs + JSON output)"
        else
            fail "verbose flag (debug logs but no JSON output)"
        fi
    else
        fail "verbose flag: expected debug output"
    fi
}

# Run standalone if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    init_standalone_test "Core Tests"
    run_core_tests
    print_summary "Core Tests"
    exit $?
fi

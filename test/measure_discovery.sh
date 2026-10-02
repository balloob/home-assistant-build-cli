#!/bin/bash
# Offline, read-only output-size measurements; optional baseline binary argument.
set -euo pipefail
HAB="${HAB:-$(cd "$(dirname "$0")/.." && pwd)/hab}"
BASELINE="${1:-$HAB}"
measure() {
    local label="$1" binary="$2" body
    shift 2
    body=$("$binary" --skip-update-check --json "$@")
    echo "$body" | jq -e '.success == true' >/dev/null
    jq -nc --arg workflow "$label" --argjson bytes "${#body}" '{workflow:$workflow,bytes:$bytes,ha_requests:0}'
}
measure baseline_full_schema "$BASELINE" schema
measure command_index "$HAB" schema --index
measure dashboard_patch_search "$HAB" schema --index --search 'dashboard patch'
measure targeted_full_schema "$HAB" schema dashboard patch
measure targeted_compact_schema "$HAB" schema dashboard patch --compact

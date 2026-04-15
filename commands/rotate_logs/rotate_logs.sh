#!/usr/bin/env bash
# ---trigger---
# name: Rotate Logs
# description: Rotate and compress logs for a service, keeping the selected retention window
# inputs:
#   - name: service
#     label: Service
#     type: closed
#     dynamic: true
#     required: true
#   - name: retention
#     label: Retention
#     type: closed
#     dynamic: true
#     required: true
# ---end---

set -euo pipefail

# get_options <input_name> <config_json>
# Prints one option per line to stdout.
#
# config keys used:
#   LOG_SERVICES        — comma-separated list of services (default: api,worker,scheduler)
#   LOG_RETENTION_OPTIONS — comma-separated retention periods  (default: 7d,14d,30d,90d)
get_options() {
    local input_name="$1"
    local config="${2:-{\}}"

    if [[ "$input_name" == "service" ]]; then
        raw=$(python3 -c "import sys,json; print(json.loads(sys.argv[1]).get('LOG_SERVICES','api,worker,scheduler'))" "$config")
        IFS=',' read -ra items <<< "$raw"
        for item in "${items[@]}"; do
            item="$(echo "$item" | xargs)"   # trim whitespace
            [[ -n "$item" ]] && echo "$item"
        done
    elif [[ "$input_name" == "retention" ]]; then
        raw=$(python3 -c "import sys,json; print(json.loads(sys.argv[1]).get('LOG_RETENTION_OPTIONS','7d,14d,30d,90d'))" "$config")
        IFS=',' read -ra items <<< "$raw"
        for item in "${items[@]}"; do
            item="$(echo "$item" | xargs)"
            [[ -n "$item" ]] && echo "$item"
        done
    fi
}

if [[ "${1:-}" == "--trigger-get-options" ]]; then
    get_options "${2:-}" "${3:-{\}}"
    exit 0
fi

# Normal execution
service=$(python3 -c "import sys,json; print(json.loads(sys.argv[1])['service'])" "$1")
retention=$(python3 -c "import sys,json; print(json.loads(sys.argv[1])['retention'])" "$1")

echo "Rotating logs for service: $service"
echo "Retention window: $retention"
echo "Done."

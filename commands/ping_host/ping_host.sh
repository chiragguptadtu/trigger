#!/usr/bin/env bash
# ---trigger---
# name: Ping Host
# description: Check if a host is reachable over the network
# inputs:
#   - name: host
#     label: Host
#     type: open
#     required: true
#   - name: count
#     label: Ping Count
#     type: closed
#     options: ["1", "3", "5"]
#     multi: false
#     required: true
# ---end---

set -euo pipefail

host=$(python3 -c "import sys,json; print(json.loads(sys.argv[1])['host'])" "$1")
count=$(python3 -c "import sys,json; print(json.loads(sys.argv[1])['count'])" "$1")

if [[ -z "$host" ]]; then
    echo "host is required" >&2
    exit 1
fi

ping -c "$count" "$host"

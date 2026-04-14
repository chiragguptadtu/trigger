# ---trigger---
# name: Sync Users
# description: Sync user records from the identity provider to the database
# inputs:
#   - name: environment
#     label: Environment
#     type: closed
#     options: [staging, production]
#     multi: false
#     required: true
#   - name: dry_run
#     label: Dry Run
#     type: closed
#     options: [yes, no]
#     multi: false
#     required: true
# ---end---

import sys
import json

def run(inputs, config):
    env = inputs.get("environment")
    dry_run = inputs.get("dry_run") == "yes"
    print(f"Syncing users on {env} (dry_run={dry_run})")
    return ""

if __name__ == "__main__":
    inputs = json.loads(sys.argv[1]) if len(sys.argv) > 1 else {}
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    result = run(inputs, config)
    if result:
        print(result, file=sys.stderr)
        sys.exit(1)

# ---trigger---
# name: Example Command
# description: A simple example command to demonstrate the platform
# inputs:
#   - name: target_env
#     label: Target Environment
#     type: closed
#     options: [staging, production]
#     multi: false
#     required: true
#   - name: reason
#     label: Reason
#     type: open
#     required: true
# ---end---

import sys
import json

def run(inputs, config):
    env = inputs.get("target_env")
    reason = inputs.get("reason")
    print(f"Running against {env}: {reason}")
    return ""

if __name__ == "__main__":
    inputs = json.loads(sys.argv[1]) if len(sys.argv) > 1 else {}
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    result = run(inputs, config)
    if result:
        print(result, file=sys.stderr)
        sys.exit(1)

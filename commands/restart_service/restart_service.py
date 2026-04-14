# ---trigger---
# name: Restart Service
# description: Restart a running service on the selected environment
# inputs:
#   - name: environment
#     label: Environment
#     type: closed
#     options: [staging, production]
#     multi: false
#     required: true
#   - name: service
#     label: Service
#     type: closed
#     options: [api, worker, scheduler, frontend]
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
    env = inputs.get("environment")
    service = inputs.get("service")
    reason = inputs.get("reason")
    print(f"Restarting {service} on {env}: {reason}")
    return ""

if __name__ == "__main__":
    inputs = json.loads(sys.argv[1]) if len(sys.argv) > 1 else {}
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    result = run(inputs, config)
    if result:
        print(result, file=sys.stderr)
        sys.exit(1)
